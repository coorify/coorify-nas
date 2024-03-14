package option

import "io/fs"

type UpdateOption struct {
	Version uint16
	EmbedFS fs.FS
}
