package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	mu            sync.Mutex
	wg            sync.WaitGroup
	version       string
	arch          string
	enabled       bool
	endpoint      string
	client        *http.Client
	cacheRoot     string
	directory     string
	installerPath string
	selected      *candidate
	info          Info
	state         State
	cancel        context.CancelFunc
	closed        bool
}

func New(version string, enabled bool) *Manager {
	cache, err := os.UserCacheDir()
	if err != nil {
		enabled = false
	}
	return &Manager{
		version: version, arch: nativeArchitecture(), enabled: enabled && runtime.GOOS == "windows" && stableVersion.MatchString(version),
		endpoint:  "https://api.github.com/repos/mibgb65-cloud/OmniCred/releases/latest",
		cacheRoot: filepath.Join(cache, "OmniCred", "updates"), state: State{Phase: "idle"},
		client: &http.Client{Timeout: 30 * time.Minute, CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("更新下载重定向次数过多")
			}
			return validateURL(request.URL.String())
		}},
	}
}

func (manager *Manager) Status() State {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	return manager.state
}

func (manager *Manager) busy() bool {
	return manager.state.Phase == "downloading" || manager.state.Phase == "verifying" || manager.state.Phase == "installing"
}

func (manager *Manager) Download() (State, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.closed || !manager.enabled || manager.selected == nil {
		return manager.state, errors.New("请先检查更新并确认有可用的安装包")
	}
	if manager.busy() || manager.state.Phase == "ready" {
		return manager.state, nil
	}
	if err := os.MkdirAll(manager.cacheRoot, 0o700); err != nil {
		return manager.state, errors.New("无法创建更新缓存目录，请检查磁盘权限")
	}
	manager.cleanOldDownloads()
	directory, err := os.MkdirTemp(manager.cacheRoot, "download-")
	if err != nil {
		return manager.state, errors.New("无法创建安装包临时文件")
	}
	manager.directory = directory
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	manager.cancel = cancel
	selected := *manager.selected
	manager.state = State{Phase: "downloading", Version: selected.version, Total: selected.asset.Size}
	manager.wg.Add(1)
	go manager.download(ctx, selected, directory)
	return manager.state, nil
}

func (manager *Manager) Cancel() State {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.cancel != nil && (manager.state.Phase == "downloading" || manager.state.Phase == "verifying") {
		manager.cancel()
	}
	return manager.state
}

func (manager *Manager) Close() {
	manager.mu.Lock()
	manager.closed = true
	if manager.cancel != nil {
		manager.cancel()
	}
	manager.mu.Unlock()
	manager.wg.Wait()
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.state.Phase != "installing" && manager.directory != "" {
		_ = os.RemoveAll(manager.directory)
	}
}

func (manager *Manager) download(ctx context.Context, selected candidate, directory string) {
	defer manager.wg.Done()
	path := filepath.Join(directory, "installer.exe.part")
	err := manager.downloadFile(ctx, selected, path)
	manager.mu.Lock()
	defer manager.mu.Unlock()
	defer manager.cancel()
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	if err == nil {
		finalPath := filepath.Join(directory, "installer.exe")
		if os.Rename(path, finalPath) != nil {
			err = errors.New("无法保存已校验的安装包")
		} else {
			manager.installerPath = finalPath
			manager.state.Phase = "ready"
			return
		}
	}
	_ = os.RemoveAll(directory)
	manager.installerPath, manager.directory = "", ""
	if errors.Is(err, context.Canceled) {
		manager.state = State{Phase: "idle"}
	} else {
		message := err.Error()
		if errors.Is(err, context.DeadlineExceeded) {
			message = "下载超时，请重试"
		}
		manager.state.Phase, manager.state.Error = "error", message
	}
}

func (manager *Manager) downloadFile(ctx context.Context, selected candidate, path string) error {
	response, err := manager.get(ctx, selected.asset.URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return downloadStatusError(response.StatusCode)
	}
	if response.ContentLength >= 0 && response.ContentLength != selected.asset.Size {
		return errors.New("安装包大小与发布清单不一致")
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return errors.New("无法写入安装包，请检查磁盘空间和权限")
	}
	defer file.Close()
	reader := io.LimitReader(response.Body, selected.asset.Size+1)
	buffer := make([]byte, 128<<10)
	var downloaded int64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		count, readErr := reader.Read(buffer)
		if count > 0 {
			downloaded += int64(count)
			if downloaded > selected.asset.Size {
				return errors.New("安装包超过预期大小，已中止下载")
			}
			if _, err := file.Write(buffer[:count]); err != nil {
				return errors.New("安装包写入失败，请检查磁盘空间")
			}
			manager.mu.Lock()
			manager.state.Downloaded = downloaded
			manager.mu.Unlock()
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return errors.New("安装包下载中断，请重试")
		}
	}
	if downloaded != selected.asset.Size {
		return errors.New("安装包下载不完整，请重试")
	}
	if err := file.Sync(); err != nil {
		return errors.New("安装包保存失败，请检查磁盘空间")
	}
	if err := file.Close(); err != nil {
		return errors.New("安装包保存失败")
	}
	manager.mu.Lock()
	manager.state.Phase = "verifying"
	manager.mu.Unlock()
	verified, err := verifyInstaller(path, selected)
	if verified != nil {
		_ = verified.Close()
	}
	return err
}

// 校验和启动之间保持只读文件锁，避免 Windows 上文件被替换或改写。
func verifyInstaller(path string, selected candidate) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() != selected.asset.Size {
		return nil, errors.New("安装包已丢失或大小不正确，请重新下载")
	}
	file, err := openInstaller(path)
	if err != nil {
		return nil, errors.New("无法读取或锁定安装包，请重新下载")
	}
	hash := sha256.New()
	count, err := io.Copy(hash, io.LimitReader(file, selected.asset.Size+1))
	if err != nil || count != selected.asset.Size || hex.EncodeToString(hash.Sum(nil)) != selected.digest {
		_ = file.Close()
		return nil, errors.New("安装包 SHA-256 校验失败，已阻止安装，请重新下载")
	}
	return file, nil
}

func (manager *Manager) Install(launch func(string) error) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.closed || !manager.enabled || manager.state.Phase != "ready" || manager.selected == nil {
		return errors.New("安装包尚未完成下载和校验")
	}
	file, err := verifyInstaller(manager.installerPath, *manager.selected)
	if err != nil {
		manager.state.Phase, manager.state.Error = "error", err.Error()
		_ = os.RemoveAll(manager.directory)
		manager.directory, manager.installerPath = "", ""
		return err
	}
	defer file.Close()
	if err := launch(manager.installerPath); err != nil {
		return err
	}
	manager.state.Phase = "installing"
	return nil
}

func (manager *Manager) cleanOldDownloads() {
	entries, _ := os.ReadDir(manager.cacheRoot)
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "download-") {
			continue
		}
		info, err := entry.Info()
		if err == nil && time.Since(info.ModTime()) > 24*time.Hour {
			_ = os.RemoveAll(filepath.Join(manager.cacheRoot, entry.Name()))
		}
	}
}
