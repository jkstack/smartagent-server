package layout

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	lapi "server/code/api"
	"server/code/client"
	"sync/atomic"

	"github.com/lwch/api"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

func (h *Handler) run(clients *client.Clients, ctx *api.Context) {
	ids := ctx.XCsv("ids")
	mode := ctx.OStr("mode", "sequence")
	errContinue := ctx.OBool("continue", false)
	user := ctx.OStr("user", "")
	pass := ctx.OStr("pass", "")
	file, _, err := ctx.File("file")
	runtime.Assert(err)

	if mode != "sequence" &&
		mode != "parallel" &&
		mode != "evenodd" {
		lapi.BadParamErr("mode")
		return
	}

	uniq := make(map[string]bool)
	for _, id := range ids {
		if uniq[id] {
			lapi.BadParamErr("ids")
			return
		}
		cli := clients.Get(id)
		if cli == nil {
			lapi.NotfoundErr(id)
			return
		}
		uniq[id] = true
	}

	idx := atomic.AddUint32(&h.idx, 1)
	runner := newRunner(h, clients, idx, ids, mode, errContinue, user, pass)
	dir, err := h.extract(file, runner.id)
	runtime.Assert(err)

	err = runner.checkAndBuild(dir, h.handlers)
	if err != nil {
		ctx.ERR(1, err.Error())
		return
	}

	h.Lock()
	h.runners[runner.id] = runner
	h.Unlock()
	go runner.run(dir)

	ctx.OK(runner.id)
}

func (h *Handler) extract(file io.Reader, taskID string) (string, error) {
	os.MkdirAll(h.cfg.CacheDir, 0755)
	dir, err := ioutil.TempDir(h.cfg.CacheDir, "layout")
	if err != nil {
		return "", err
	}
	tr := tar.NewReader(file)
	write := func(dir string, r io.Reader) error {
		f, err := os.Create(dir)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, r)
		return err
	}
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return dir, nil
			}
			os.RemoveAll(dir)
			return "", err
		}
		if hdr.Typeflag == tar.TypeDir {
			continue
		}
		if hdr.Typeflag != tar.TypeReg {
			logging.Info("skip file: %s", hdr.Name)
			continue
		}
		target := filepath.Join(dir, hdr.Name)
		os.MkdirAll(filepath.Dir(target), 0755)
		err = write(target, tr)
		if err != nil {
			os.RemoveAll(dir)
			return "", err
		}
	}
}
