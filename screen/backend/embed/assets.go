package assets

import "embed"

const Version = uint16(0x0005)

//go:embed bootloader.bin nas-ui.bin partition-table.bin stub/*
var FS embed.FS
