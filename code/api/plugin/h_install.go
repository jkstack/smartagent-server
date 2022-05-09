package plugin

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path"
	"server/code/client"

	"github.com/lwch/api"
	"github.com/lwch/runtime"
)

func (h *Handler) install(clients *client.Clients, ctx *api.Context) {
	name := ctx.XStr("name")
	version := ctx.XStr("version")
	file, _, err := ctx.File("file")
	runtime.Assert(err)
	defer file.Close()
	dir := path.Join(h.cfg.PluginDir, name, version)
	runtime.Assert(os.MkdirAll(dir, 0755))
	gr, err := gzip.NewReader(file)
	runtime.Assert(err)
	defer gr.Close()
	write := func(dir string, r io.Reader) {
		f, err := os.Create(dir)
		runtime.Assert(err)
		defer f.Close()
		_, err = io.Copy(f, r)
		runtime.Assert(err)
	}
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		write(path.Join(dir, hdr.Name), tr)
	}

	h.cfg.LoadPlugin()

	ctx.OK(nil)
}
