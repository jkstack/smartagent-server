package file

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	lapi "server/code/api"
	"server/code/client"
	"server/code/utils"
	"time"

	"github.com/jkstack/anet"
	"github.com/lwch/api"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

func (h *Handler) download(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")
	dir := ctx.XStr("dir")
	timeout := ctx.OInt("timeout", 600)

	cli := clients.Get(id)
	if cli == nil {
		ctx.NotFound("client")
	}

	p := h.cfg.GetPlugin("file", cli.OS(), cli.Arch())
	if p == nil {
		lapi.PluginNotInstalledErr("file")
	}

	taskID, err := cli.SendDownload(p, dir)
	runtime.Assert(err)
	defer cli.ChanClose(taskID)

	h.stUsage.Inc()

	logging.Info("download [%s] on %s, task_id=%s, plugin.version=%s", dir, id, taskID, p.Version)

	var rep *anet.Msg
	select {
	case rep = <-cli.ChanRead(taskID):
	case <-time.After(api.RequestTimeout):
		ctx.Timeout()
	}

	switch {
	case rep.Type == anet.TypeError:
		ctx.ERR(http.StatusServiceUnavailable, rep.ErrorMsg)
		return
	case rep.Type != anet.TypeDownloadRep:
		ctx.ERR(http.StatusInternalServerError, fmt.Sprintf("invalid message type: %d", rep.Type))
		return
	}

	if !rep.DownloadRep.OK {
		logging.Error("download [%s] on %s failed, task_id=%s, msg=%s", dir, id, taskID, rep.DownloadRep.ErrMsg)
		ctx.HTTPServiceUnavailable(rep.DownloadRep.ErrMsg)
		return
	}
	logging.Info("download [%s] on %s success, task_id=%s, size=%d, md5=%x", dir, id, taskID,
		rep.DownloadRep.Size, rep.DownloadRep.MD5)

	f, err := tmpFile(ctx, h.cfg.CacheDir, rep.DownloadRep.Size)
	if err != nil {
		logging.Error("download [%s] on %s failed, task_id=%s, err=%v", dir, id, taskID, err)
		ctx.HTTPServiceUnavailable(err.Error())
		return
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	left := rep.DownloadRep.Size
	after := time.After(time.Duration(timeout) * time.Second)
	for {
		var msg *anet.Msg
		select {
		case msg = <-cli.ChanRead(taskID):
		case <-after:
			ctx.HTTPTimeout()
			return
		}
		switch msg.Type {
		case anet.TypeDownloadData:
			n, err := writeFile(f, msg.DownloadData)
			if err != nil {
				logging.Error("download [%s] on %s, task_id=%s, err=%v", dir, id, taskID, err)
				ctx.HTTPServiceUnavailable(err.Error())
				return
			}
			left -= uint64(n)
			if left == 0 {
				serveFile(ctx, f, id, dir, taskID, rep.DownloadRep.MD5)
				return
			}
		case anet.TypeDownloadError:
			ctx.HTTPServiceUnavailable(msg.DownloadError.Msg)
			return
		}
	}
}

func tmpFile(ctx *api.Context, cacheDir string, size uint64) (*os.File, error) {
	tmp := path.Join(cacheDir, "download")
	os.MkdirAll(tmp, 0755)
	f, err := os.CreateTemp(tmp, "dl")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = fillFile(f, size)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}
	f.Close()
	f, err = os.OpenFile(f.Name(), os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}
	return f, err
}

func serveFile(ctx *api.Context, f *os.File, id, dir, taskID string, src [md5.Size]byte) {
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		logging.Error("download [%s] on %s, task_id=%s, err=%v", dir, id, taskID, err)
		ctx.HTTPServiceUnavailable(err.Error())
		return
	}
	dst, err := utils.MD5From(f)
	if err != nil {
		logging.Error("download [%s] on %s, task_id=%s, err=%v", dir, id, taskID, err)
		ctx.HTTPServiceUnavailable(err.Error())
		return
	}
	if !bytes.Equal(dst[:], src[:]) {
		logging.Error("download [%s] on %s, task_id=%s, invalid md5checksum, src=%x, dst=%x",
			dir, id, taskID, src, dst)
		ctx.HTTPConflict("invalid checksum")
		return
	}
	f.Close()
	ctx.ServeFile(f.Name())
}
