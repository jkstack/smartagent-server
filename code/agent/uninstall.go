package agent

import (
	"context"
	"fmt"
	"server/code/sshcli"
)

func Uninstall(cli *sshcli.Client, pass string) (string, error) {
	version, err := cli.Do(context.Background(), "cat "+DefaultInstallDir+"/.version", "")
	if err != nil {
		return "", fmt.Errorf("获取版本号失败：%v", err)
	}

	_, err = sudo(cli, "systemctl stop smartagent", pass)
	if err != nil {
		_, err = sudo(cli, "service smartagent stop", pass)
	}
	if err != nil {
		sudo(cli, "/etc/init.d/smartagent stop", pass)
	}

	sudo(cli, "systemctl disable smartagent", pass)
	sudo(cli, "update-rc.d smartagent remove", pass)

	sudo(cli, DefaultInstallDir+"/bin/smartagent -conf "+
		DefaultInstallDir+"/conf/client.conf -action uninstall", pass)

	sudo(cli, "rm -fr "+DefaultInstallDir, pass)

	return version, nil
}
