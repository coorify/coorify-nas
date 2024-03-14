package main

import (
	"embed"
	"os"
	"os/signal"
	"syscall"

	"github.com/coorify/be/device"
	"github.com/coorify/be/firmeware"
	"github.com/coorify/be/monitor"
	"github.com/coorify/be/openwrt"
	"github.com/coorify/be/option"
	"github.com/jinzhu/configor"
	_ "github.com/joho/godotenv/autoload"
)

const Version = uint16(0x0005)

// bootloader.bin nas-ui.bin partition-table.bin stub/*

//go:embed embed/*
var embedFS embed.FS

func load(opt *option.Option) error {
	loader := configor.New(&configor.Config{})

	files := os.Getenv("CONFIG_FILE")
	if files == "" {
		files = "config.yml"
	}

	return loader.Load(opt, files)
}

func main() {
	o := &option.Option{}
	if err := load(o); err != nil {
		panic(err)
	}

	wrt := openwrt.NewClient(&o.OpenWrt)
	if err := wrt.Sigin(); err != nil {
		panic(err)
	}

	uo := &option.UpdateOption{
		Version: uint16(0x0005),
		EmbedFS: embedFS,
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT)

	name := device.WaitPort()
	drv := device.NewDriver(name)

	if err := firmeware.Update(drv, uo); err != nil {
		panic(err)
	}

	mtr := monitor.NewMonitor(drv, wrt)
	if err := mtr.Start(); err != nil {
		panic(err)
	}

	<-sigint

	if err := mtr.Stop(); err != nil {
		panic(err)
	}
}
