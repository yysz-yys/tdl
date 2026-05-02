package main

import (
	"context"
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"go.uber.org/zap"

	"github.com/iyear/tdl/app/daemon"
)

// Wails 会在这个目录查找前端静态资源
// 由于 wails.json 中的 build 命令会自动将 dashboard/dist 复制到特定位置，
// 但在简单的多模块结构中，我们也可以使用 //go:embed 包含构建后的产物。
// 为了简化 GitHub Actions 打包，我们假设外层构建脚本已经把 dist 移入到了当前目录下的 frontend/dist
//go:embed frontend/dist
var assets embed.FS

type App struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	c, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	logger, _ := zap.NewProduction()
	// 在后台启动守护进程，供前端 Axios 和 WebSocket 调用
	srv := daemon.NewServer(8080, logger)
	go srv.Start(c)
}

func (a *App) shutdown(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
}

func main() {
	// 如果需要设置环境变量，可以在此处配置
	os.Setenv("TDL_ENV", "desktop")

	app := &App{}
	err := wails.Run(&options.App{
		Title:  "TDL Dashboard",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
	})

	if err != nil {
		log.Fatal(err)
	}
}
