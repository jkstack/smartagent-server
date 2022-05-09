package agent

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"server/code/sshcli"
	"server/code/utils"

	"github.com/lwch/kvconf"
	"github.com/lwch/runtime"
)

const DefaultInstallDir = "/opt/smartagent"

func Extract(cli *sshcli.Client, pass string, data []byte) error {
	_, err := cli.Stat(DefaultInstallDir + "/.success")
	if !os.IsNotExist(err) {
		return errors.New("无法重复安装")
	}
	err = cli.UploadFrom(bytes.NewReader(data), "smartagent.tar.gz")
	if err != nil {
		return fmt.Errorf("文件上传失败：%v", err)
	}
	defer cli.Remove("smartagent.tar.gz")
	str, err := sudo(cli, "tar -xhf smartagent.tar.gz -C /", pass)
	if err != nil {
		return fmt.Errorf("安装包解压失败：%s", str)
	}
	return nil
}

type Config struct {
	ID        string      `kv:"id"`
	Server    string      `kv:"server"`
	User      string      `kv:"user"`
	PluginDir string      `kv:"plugin_dir"`
	LogDir    string      `kv:"log_dir"`
	LogSize   utils.Bytes `kv:"log_size"`
	LogRotate int         `kv:"log_rotate"`
	CPU       uint32      `kv:"cpu_limit"`
	Memory    utils.Bytes `kv:"memory_limit"`
}

func Install(cli *sshcli.Client, pass string, cfg Config) (string, error) {
	if !path.IsAbs(cfg.PluginDir) {
		cfg.PluginDir = DefaultInstallDir + "/" + cfg.PluginDir
	}
	if !path.IsAbs(cfg.LogDir) {
		cfg.LogDir = DefaultInstallDir + "/" + cfg.LogDir
	}

	var buf bytes.Buffer
	runtime.Assert(kvconf.NewEncoder(&buf).Encode(cfg))
	content := buf.String()

	err := cli.UploadFrom(&buf, "client.conf")
	if err != nil {
		return "", fmt.Errorf("生成配置文件失败：%v", err)
	}
	defer cli.Remove("client.conf")
	str, err := sudo(cli, "mv client.conf "+DefaultInstallDir+"/conf/client.conf", pass)
	if err != nil {
		return "", fmt.Errorf("移动配置文件失败：%s", str)
	}

	sudo(cli, DefaultInstallDir+"/bin/smartagent -conf "+
		DefaultInstallDir+"/conf/client.conf -action install", pass)
	sudo(cli, "systemctl enable smartagent", pass)
	sudo(cli, "update-rc.d smartagent defaults", pass)

	str, err = sudo(cli, "systemctl restart smartagent", pass)
	if err != nil {
		str, err = sudo(cli, "service smartagent restart", pass)
	}
	if err != nil {
		str, err = sudo(cli, "/etc/init.d/smartagent restart", pass)
	}
	if err != nil {
		return "", fmt.Errorf("启动失败：%s", str)
	}

	sudo(cli, "touch "+DefaultInstallDir+"/.success", pass)

	return content, nil
}
