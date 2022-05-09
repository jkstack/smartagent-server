package client

import (
	"crypto/md5"
	"os"
	"server/code/conf"
	"server/code/utils"

	"github.com/jkstack/anet"
)

type UploadContext struct {
	Dir      string
	Name     string
	Auth     string
	User     string
	Pass     string
	Mod      int
	OwnUser  string
	OwnGroup string
	Size     uint64
	Md5      [md5.Size]byte
	Md5Check bool
	// send data
	Data []byte
	// send file
	Uri   string
	Token string
}

func (cli *Client) SendLS(p *conf.PluginInfo, dir string) (string, error) {
	id, err := utils.TaskID()
	if err != nil {
		return "", err
	}
	var msg anet.Msg
	msg.Type = anet.TypeLsReq
	msg.TaskID = id
	msg.Plugin = fillPlugin(p)
	msg.LSReq = &anet.LsReq{Dir: dir}
	cli.Lock()
	cli.taskRead[id] = make(chan *anet.Msg)
	cli.Unlock()
	cli.chWrite <- &msg
	return id, nil
}

func (cli *Client) SendDownload(p *conf.PluginInfo, dir string) (string, error) {
	id, err := utils.TaskID()
	if err != nil {
		return "", err
	}
	var msg anet.Msg
	msg.Type = anet.TypeDownloadReq
	msg.TaskID = id
	msg.Plugin = fillPlugin(p)
	msg.DownloadReq = &anet.DownloadReq{Dir: dir}
	cli.Lock()
	cli.taskRead[id] = make(chan *anet.Msg)
	cli.Unlock()
	cli.chWrite <- &msg
	return id, nil
}

func (cli *Client) SendUpload(p *conf.PluginInfo, ctx UploadContext, id string) (string, error) {
	if len(id) == 0 {
		var err error
		id, err = utils.TaskID()
		if err != nil {
			return "", err
		}
	}
	var msg anet.Msg
	msg.Type = anet.TypeUpload
	msg.TaskID = id
	msg.Plugin = fillPlugin(p)
	msg.Upload = &anet.Upload{
		Dir:      ctx.Dir,
		Name:     ctx.Name,
		Auth:     ctx.Auth,
		User:     ctx.User,
		Pass:     ctx.Pass,
		Mod:      os.FileMode(ctx.Mod),
		OwnUser:  ctx.OwnUser,
		OwnGroup: ctx.OwnGroup,
		Size:     ctx.Size,
		MD5:      ctx.Md5,
	}
	if len(ctx.Data) > 0 {
		msg.Upload.Data = utils.EncodeData(ctx.Data)
	} else if len(ctx.Uri) > 0 {
		msg.Upload.URI = ctx.Uri
		msg.Upload.Token = ctx.Token
	}
	cli.Lock()
	cli.taskRead[id] = make(chan *anet.Msg)
	cli.Unlock()
	cli.chWrite <- &msg
	return id, nil
}
