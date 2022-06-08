package layout

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"server/code/client"
	"server/code/conf"
	"server/code/utils"
	"time"

	"github.com/jkstack/anet"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

type fileHandler struct {
	taskInfo
	push bool
	src  string
	dst  string
}

func (h *fileHandler) Check(t Task) error {
	if t.Action != "push" && t.Action != "pull" {
		return fmt.Errorf("unexpected action %s for task [%s], supported push or pull",
			t.Action, t.Name)
	}
	if len(t.Src) == 0 {
		return fmt.Errorf("missing src for task [%s]", t.Name)
	}
	if t.Action == "pull" && !filepath.IsAbs(t.Src) {
		return fmt.Errorf("unexpected absolute path in src for task [%s]", t.Name)
	}
	if len(t.Dst) == 0 {
		return fmt.Errorf("missing dst for task [%s]", t.Name)
	}
	if !filepath.IsAbs(t.Dst) {
		return fmt.Errorf("unexpected relative path in dst for task [%s]", t.Name)
	}
	return nil
}

func (h *fileHandler) Clone(t Task, info taskInfo) taskHandler {
	return &fileHandler{
		taskInfo: info,
		push:     t.Action == "push",
		src:      t.Src,
		dst:      t.Dst,
	}
}

func (h *fileHandler) Run(id, dir, user, pass string, args map[string]string) error {
	deadline := h.deadline()
	args["DEADLINE"] = fmt.Sprintf("%d", deadline.Unix())
	action := "push"
	if !h.push {
		action = "pull"
	}
	logging.Info("%s file on agent %s", action, id)
	cli := h.parent.GetClient(id)
	if cli == nil {
		return errClientNotfound(id)
	}
	p := h.parent.GetPlugin("file", cli.OS(), cli.Arch())
	if p == nil {
		return errPluginNotInstalled("file")
	}
	if h.push {
		return h.upload(cli, p, dir, user, pass, deadline)
	}
	return h.download(cli, p, user, pass, deadline)
}

func (h *fileHandler) upload(cli *client.Client, p *conf.PluginInfo,
	dir, user, pass string, deadline time.Time) error {
	src := h.src
	if !filepath.IsAbs(src) {
		src = filepath.Join(dir, src)
	}
	token, err := runtime.UUID(16, "0123456789abcdef")
	if err != nil {
		return fmt.Errorf("generate token: %v", err)
	}
	taskID, err := utils.TaskID()
	if err != nil {
		return fmt.Errorf("generate task_id: %v", err)
	}
	md5, err := utils.MD5Checksum(src)
	if err != nil {
		return fmt.Errorf("calculate md5 checksum: %v", err)
	}
	fi, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("get file info: %v", err)
	}
	uri := "/file/upload/" + taskID
	h.parent.parent.fh.LogUploadCache(taskID, src, token, deadline, false)
	defer h.parent.parent.fh.RemoveUploadCache(taskID)
	taskID, err = cli.SendUpload(p, client.UploadContext{
		Dir:  filepath.Dir(h.dst),
		Name: filepath.Base(h.dst),
		Auth: h.auth,
		User: user,
		Pass: pass,
		Mod:  int(fi.Mode()),
		// OwnUser:  ownUser,
		// OwnGroup: ownGroup,
		Size:  uint64(fi.Size()),
		Md5:   md5,
		Uri:   uri,
		Token: token,
	}, taskID)
	if err != nil {
		return err
	}

	h.parent.parent.stFileUsage.Inc()
	h.parent.parent.stTotalTasks.Inc()

	var rep *anet.Msg
	select {
	case rep = <-cli.ChanRead(taskID):
	case <-time.After(deadline.Sub(time.Now())):
		return errTimeout
	}

	switch {
	case rep.Type == anet.TypeError:
		return errors.New(rep.ErrorMsg)
	case rep.Type != anet.TypeUploadRep:
		return fmt.Errorf("invalid message type: %d", rep.Type)
	}

	if !rep.UploadRep.OK {
		return errors.New(rep.UploadRep.ErrMsg)
	}

	return nil
}

func (h *fileHandler) download(cli *client.Client, p *conf.PluginInfo,
	user, pass string, deadline time.Time) error {
	taskID, err := cli.SendDownload(p, h.src)
	if err != nil {
		return err
	}
	defer cli.ChanClose(taskID)

	h.parent.parent.stFileUsage.Inc()
	h.parent.parent.stTotalTasks.Inc()

	var rep *anet.Msg
	after := time.After(deadline.Sub(time.Now()))
	select {
	case rep = <-cli.ChanRead(taskID):
	case <-after:
		return errTimeout
	}

	switch {
	case rep.Type == anet.TypeError:
		return errors.New(rep.ErrorMsg)
	case rep.Type != anet.TypeDownloadRep:
		return fmt.Errorf("invalid message type: %d", rep.Type)
	}

	if !rep.DownloadRep.OK {
		return errors.New(rep.DownloadRep.ErrMsg)
	}

	f, err := os.Create(h.dst)
	if err != nil {
		return fmt.Errorf("create destination file: %v", err)
	}
	defer f.Close()

	left := rep.DownloadRep.Size
	for {
		var msg *anet.Msg
		select {
		case msg = <-cli.ChanRead(taskID):
		case <-after:
			return errTimeout
		}
		switch msg.Type {
		case anet.TypeDownloadData:
			n, err := writeFile(f, msg.DownloadData)
			if err != nil {
				return fmt.Errorf("write data: %v", err)
			}
			left -= uint64(n)
			if left == 0 {
				return nil
			}
		case anet.TypeDownloadError:
			return errors.New(msg.DownloadError.Msg)
		}
	}
}

func writeFile(f *os.File, data *anet.DownloadData) (int, error) {
	_, err := f.Seek(int64(data.Offset), io.SeekStart)
	if err != nil {
		return 0, err
	}
	dec, err := utils.DecodeData(data.Data)
	if err != nil {
		return 0, err
	}
	return f.Write(dec)
}
