package agent

import (
	"context"
	"fmt"
	"server/code/sshcli"
	"strings"
)

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
