package agent

import (
	"context"
	"fmt"
	"server/code/sshcli"
	"strings"

	"github.com/lwch/logging"
)

func service(cli *sshcli.Client, pass, name, action string) (string, error) {
	_, err := sudo(cli, "systemctl "+action+" "+name, pass)
	if err == nil {
		return "", nil
	}
	logging.Info("systemctl %s %s failed: %v", action, name, err)
	_, err = sudo(cli, "service "+name+" "+action, pass)
	if err == nil {
		return "", nil
	}
	logging.Info("service %s %s failed: %v", name, action, err)
	str, err := sudo(cli, "/etc/init.d/"+name+" "+action, pass)
	if err == nil {
		return "", nil
	}
	logging.Info("/etc/init.d/%s %s failed: %v", name, action, err)
	return str, err
}

func sudo(cli *sshcli.Client, cmd, pass string) (string, error) {
	str, err := cli.Do(context.Background(), "sudo -k", "")
	if err != nil {
		return fmt.Sprintf("重设sudo密码失败：%s", str), err
	}
	str, err = cli.Do(context.Background(), "sudo -S -b "+cmd+" && sleep 1", pass+"\n")
	if err != nil {
		return str, err
	}
	idx := strings.Index(str, "\r\n")
	if idx != -1 {
		str = str[idx+2:]
	}
	return str, err
}
