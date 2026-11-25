package main

import (
	"embed"
	"log"
	"time"

	"github.com/maxachis/clean-tracker/wails/service"
	"github.com/maxachis/clean-tracker/wails/store"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

const (
	windowTitle  = "Clean Tracker"
	windowWidth  = 900
	windowHeight = 700
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	dbPath, err := store.DefaultDBPath()
	if err != nil {
		log.Fatalf("resolve db path: %v", err)
	}
	s, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	app := NewApp(service.New(s, time.Now))

	err = wails.Run(&options.App{
		Title:            windowTitle,
		Width:            windowWidth,
		Height:           windowHeight,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind:             []interface{}{app},
	})
	if err != nil {
		log.Fatalf("wails run: %v", err)
	}
}
