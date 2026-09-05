package updater

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var stableVersion = regexp.MustCompile(`^v?(0|[1-9][0-9]{0,8})\.(0|[1-9][0-9]{0,8})\.(0|[1-9][0-9]{0,8})$`)

func (manager *Manager) Check(ctx context.Context) (Info, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.busy() || manager.state.Phase == "ready" {
		return manager.info, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	var latest release
	status, err := manager.readJSON(ctx, manager.endpoint, 1<<20, &latest)
	if err != nil {
		return Info{}, err
	}
	info := Info{CurrentVersion: manager.version, LatestVersion: manager.version,
		ReleaseURL: RepositoryURL + "/releases", CheckedAt: time.Now().UTC(), Status: "no_releases"}
	if status == http.StatusNotFound {
		manager.info, manager.selected = info, nil
		return info, nil
	}
	if status != http.StatusOK || !stableVersion.MatchString(latest.Tag) || latest.Draft || latest.Prerelease {
		return Info{}, errors.New("无法读取稳定版本信息，请稍后重试")
	}
	info.LatestVersion, info.Status = latest.Tag, "ok"
	info.ReleaseURL = RepositoryURL + "/releases/tag/" + latest.Tag
	info.PublishedAt = &latest.PublishedAt
	info.UpdateAvailable = compareVersion(latest.Tag, manager.version) > 0
	manager.selected = nil
	if info.UpdateAvailable {
		if !manager.enabled {
			info.UnavailableReason = "应用内更新仅支持 Windows 正式安装版；当前版本请从发布页安装。"
		} else {
			selected, err := manager.selectInstaller(ctx, latest)
			if err != nil {
				info.UnavailableReason = err.Error()
			} else {
				manager.selected = selected
				info.DownloadAvailable = true
			}
		}
	}
	manager.info = info
	return info, nil
}

func (manager *Manager) selectInstaller(ctx context.Context, latest release) (*candidate, error) {
	manifestAsset, err := findAsset(latest, "update-manifest.json")
	if err != nil {
		return nil, errors.New("此版本未提供应用内更新清单，请从发布页安装")
	}
	if manifestAsset.Size <= 0 || manifestAsset.Size > 64<<10 {
		return nil, errors.New("更新清单大小无效")
	}
	var data manifest
	status, err := manager.readJSON(ctx, manifestAsset.URL, 64<<10, &data)
	if err != nil || status != http.StatusOK {
		return nil, errors.New("无法获取更新校验清单，请重新检查更新")
	}
	if data.Protocol != 1 || data.Version != latest.Tag || len(data.Installers) > 8 {
		return nil, errors.New("更新清单与版本不匹配或更新协议不受支持")
	}
	var selected *candidate
	for _, installer := range data.Installers {
		if installer.OS != "windows" || installer.Arch != manager.arch {
			continue
		}
		name := "OmniCred-" + manager.arch + "-installer.exe"
		digest, err := hex.DecodeString(installer.SHA256)
		if selected != nil || installer.Name != name || err != nil || len(digest) != 32 || installer.Size <= 0 || installer.Size > maxInstallerSize {
			return nil, errors.New("安装包校验信息无效")
		}
		asset, err := findAsset(latest, name)
		if err != nil || asset.Size != installer.Size {
			return nil, errors.New("安装包与更新清单不匹配")
		}
		selected = &candidate{version: latest.Tag, asset: asset, digest: strings.ToLower(installer.SHA256)}
	}
	if selected == nil {
		return nil, errors.New("此版本没有适配当前系统架构的安装包")
	}
	return selected, nil
}

func findAsset(latest release, name string) (releaseAsset, error) {
	var found *releaseAsset
	for _, asset := range latest.Assets {
		if asset.Name != name {
			continue
		}
		expected := RepositoryURL + "/releases/download/" + latest.Tag + "/" + name
		if found != nil || asset.URL != expected {
			return releaseAsset{}, errors.New("发布附件来源或名称无效")
		}
		found = &asset
	}
	if found == nil {
		return releaseAsset{}, errors.New("发布附件缺失")
	}
	return *found, nil
}

func (manager *Manager) readJSON(ctx context.Context, address string, limit int64, target any) (int, error) {
	response, err := manager.get(ctx, address)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return response.StatusCode, nil
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || int64(len(content)) > limit {
		return 0, errors.New("更新信息读取失败或超过大小限制")
	}
	if err := json.Unmarshal(content, target); err != nil {
		return 0, errors.New("更新信息格式无效")
	}
	return response.StatusCode, nil
}

func (manager *Manager) get(ctx context.Context, address string) (*http.Response, error) {
	if err := validateURL(address); err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, errors.New("更新请求无效")
	}
	request.Header.Set("User-Agent", "OmniCred/"+manager.version)
	request.Header.Set("Accept", "application/octet-stream")
	if request.URL.Host == "api.github.com" {
		request.Header.Set("Accept", "application/vnd.github+json")
	}
	response, err := manager.client.Do(request)
	if err != nil {
		// 重定向 URL 可能含临时签名，不能将底层网络错误直接展示或记录。
		return nil, errors.New("无法连接更新服务器，请检查网络后重试")
	}
	return response, nil
}

func validateURL(address string) error {
	parsed, err := url.Parse(address)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Fragment != "" || parsed.Port() != "" {
		return errors.New("更新下载地址无效")
	}
	switch parsed.Host {
	case "api.github.com":
		if parsed.Path == "/repos/mibgb65-cloud/OmniCred/releases/latest" {
			return nil
		}
	case "github.com":
		if strings.HasPrefix(parsed.Path, "/mibgb65-cloud/OmniCred/releases/download/") {
			return nil
		}
	case "release-assets.githubusercontent.com", "objects.githubusercontent.com":
		return nil
	}
	return errors.New("更新下载地址不属于受信任的发布源")
}

func compareVersion(left, right string) int {
	parse := func(value string) [3]int {
		var result [3]int
		parts := strings.Split(strings.TrimPrefix(value, "v"), ".")
		for index := 0; index < len(parts) && index < 3; index++ {
			result[index], _ = strconv.Atoi(parts[index])
		}
		return result
	}
	a, b := parse(left), parse(right)
	for index := range a {
		if a[index] > b[index] {
			return 1
		}
		if a[index] < b[index] {
			return -1
		}
	}
	return 0
}

func downloadStatusError(status int) error {
	return fmt.Errorf("下载安装包失败（HTTP %d），请重试", status)
}
