package file

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
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

const uploadLimit = 1024 * 1024

type uploadInfo struct {
	token   string
	dir     string
	rm      bool
	timeout time.Time
}

func (h *Handler) upload(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")
	dir := ctx.XStr("dir")
	enc := ctx.OStr("md5", "")
	auth := ctx.OStr("auth", "")
	user := ctx.OStr("user", "")
	pass := ctx.OStr("pass", "")
	mod := ctx.OInt("mod", 0644)
	ownUser := ctx.OStr("own_user", "")
	ownGroup := ctx.OStr("own_group", "")
	file, hdr, err := ctx.File("file")
	runtime.Assert(err)
	timeout := ctx.OInt("timeout", 60)

	var srcMD5 [md5.Size]byte
	if len(enc) > 0 {
		encData, err := hex.DecodeString(enc)
		runtime.Assert(err)

		copy(srcMD5[:], encData)
	}

	cli := clients.Get(id)
	if cli == nil {
		ctx.NotFound("client")
	}

	p := h.cfg.GetPlugin("file", cli.OS(), cli.Arch())
	if p == nil {
		lapi.PluginNotInstalledErr("file")
	}

	var taskID string
	if hdr.Size <= uploadLimit {
		var data []byte
		data, err = ioutil.ReadAll(file)
		runtime.Assert(err)
		dstMD5 := md5.Sum(data)
		info := client.UploadContext{
			Dir:      dir,
			Name:     hdr.Filename,
			Auth:     auth,
			User:     user,
			Pass:     pass,
			Mod:      mod,
			OwnUser:  ownUser,
			OwnGroup: ownGroup,
			Size:     uint64(hdr.Size),
			Data:     data,
		}
		if len(enc) > 0 {
			if !bytes.Equal(srcMD5[:], dstMD5[:]) {
				ctx.ERR(2, "invalid checksum")
				return
			}
			info.Md5 = srcMD5
		} else {
			info.Md5 = md5.Sum(data)
		}
		taskID, err = cli.SendUpload(p, info, "")
	} else {
		var tmpDir string
		var dstMD5 [md5.Size]byte
		tmpDir, dstMD5, err = dumpFile(file, path.Join(h.cfg.CacheDir, "upload"))
		runtime.Assert(err)
		defer os.Remove(tmpDir)
		if len(enc) > 0 && !bytes.Equal(srcMD5[:], dstMD5[:]) {
			ctx.ERR(2, "invalid checksum")
			return
		}
		token, err := runtime.UUID(16, "0123456789abcdef")
		runtime.Assert(err)
		taskID, err = utils.TaskID()
		runtime.Assert(err)
		uri := "/file/upload/" + taskID
		h.LogUploadCache(taskID, tmpDir, token,
			time.Now().Add(time.Duration(timeout)*time.Second), true)
		defer h.RemoveUploadCache(taskID)
		taskID, err = cli.SendUpload(p, client.UploadContext{
			Dir:      dir,
			Name:     hdr.Filename,
			Auth:     auth,
			User:     user,
			Pass:     pass,
			Mod:      mod,
			OwnUser:  ownUser,
			OwnGroup: ownGroup,
			Size:     uint64(hdr.Size),
			Md5:      dstMD5,
			Uri:      uri,
			Token:    token,
		}, taskID)
	}
	runtime.Assert(err)
	defer cli.ChanClose(taskID)

	logging.Info("upload [%s] to %s on %s, task_id=%s, plugin.version=%s",
		hdr.Filename, dir, id, taskID, p.Version)

	var rep *anet.Msg
	select {
	case rep = <-cli.ChanRead(taskID):
	case <-time.After(time.Duration(timeout) * time.Second):
		ctx.Timeout()
	}

	switch {
	case rep.Type == anet.TypeError:
		ctx.ERR(http.StatusServiceUnavailable, rep.ErrorMsg)
		return
	case rep.Type != anet.TypeUploadRep:
		ctx.ERR(http.StatusInternalServerError, fmt.Sprintf("invalid message type: %d", rep.Type))
		return
	}

	if !rep.UploadRep.OK {
		ctx.ERR(1, rep.UploadRep.ErrMsg)
		return
	}

	ctx.OK(nil)
}

func dumpFile(f multipart.File, dir string) (string, [md5.Size]byte, error) {
	var md [md5.Size]byte
	os.MkdirAll(dir, 0755)
	dst, err := os.CreateTemp(dir, "ul")
	if err != nil {
		return "", md, err
	}
	defer dst.Close()
	enc := md5.New()
	_, err = io.Copy(io.MultiWriter(dst, enc), f)
	if err != nil {
		return "", md, err
	}
	copy(md[:], enc.Sum(nil))
	return dst.Name(), md, nil
}
