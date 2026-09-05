package desktop

import (
	"errors"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"omnicred/internal/updater"
)

func (app *App) UpdateStatus() updater.State {
	if app.updater == nil {
		return updater.State{Phase: "idle"}
	}
	return app.updater.Status()
}

func (app *App) DownloadUpdate() (updater.State, error) {
	if app.updater == nil {
		return updater.State{}, errors.New("当前环境不支持应用内更新")
	}
	return app.updater.Download()
}

func (app *App) CancelUpdate() updater.State {
	if app.updater == nil {
		return updater.State{Phase: "idle"}
	}
	return app.updater.Cancel()
}

func (app *App) RestartToUpdate() error {
	if app.updater == nil {
		return errors.New("当前环境不支持应用内更新")
	}
	app.mu.RLock()
	ctx := app.ctx
	app.mu.RUnlock()
	if ctx == nil {
		return errors.New("桌面应用尚未就绪")
	}
	if err := app.updater.Install(app.startInstaller); err != nil {
		return err
	}
	// 安装器已通过 UAC 启动并负责等待进程退出；先返回调用结果再关闭窗口。
	go func() {
		time.Sleep(300 * time.Millisecond)
		runtime.Quit(ctx)
	}()
	return nil
}
