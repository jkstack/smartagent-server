package agent

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"server/code/sshcli"
	"strings"

	"github.com/lwch/kvconf"
	"github.com/lwch/runtime"
)

func Upgrade(cli *sshcli.Client, pass string, data []byte, restart bool) (string, error) {
	old, err := cli.Do(context.Background(), "cat "+
		path.Join(DefaultInstallDir, ".version"), "")
	runtime.Assert(err)

	confDir := path.Join(DefaultInstallDir, "conf", "client.conf")
	conf, err := cli.Do(context.Background(), "cat "+confDir, "")
	runtime.Assert(err)

	var src map[string]string
	runtime.Assert(kvconf.NewDecoder(strings.NewReader(conf)).Decode(&src))

	err = cli.UploadFrom(bytes.NewReader(data), "smartagent.tar.gz")
	if err != nil {
		return "", fmt.Errorf("文件上传失败：%v", err)
	}
	defer cli.Remove("smartagent.tar.gz")
	str, err := sudo(cli, "tar -xhf smartagent.tar.gz -C /", pass)
	if err != nil {
		return "", fmt.Errorf("安装包解压失败：%s", str)
	}

	conf, err = cli.Do(context.Background(), "cat "+confDir, "")
	runtime.Assert(err)

	var dst map[string]string
	runtime.Assert(kvconf.NewDecoder(strings.NewReader(conf)).Decode(&dst))

	for k, v := range dst {
		if _, ok := src[k]; !ok {
			src[k] = v
		}
	}

	var buf bytes.Buffer
	runtime.Assert(kvconf.NewEncoder(&buf).Encode(src))

	err = cli.UploadFrom(&buf, "client.conf")
	if err != nil {
		return "", fmt.Errorf("生成配置文件失败：%v", err)
	}
	defer cli.Remove("client.conf")
	str, err = sudo(cli, "mv client.conf "+DefaultInstallDir+"/conf/client.conf", pass)
	if err != nil {
		return "", fmt.Errorf("移动配置文件失败：%s", str)
	}

	if restart {
		str, err = sudo(cli, "systemctl restart smartagent", pass)
		if err != nil {
			str, err = sudo(cli, "service smartagent restart", pass)
		}
		if err != nil {
			str, err = sudo(cli, "/etc/init.d/smartagent restart", pass)
		}
		if err != nil {
			return "", fmt.Errorf("重启失败：%s", str)
		}
	}

	return old, nil
}
