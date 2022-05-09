package batch

import (
	"os"
	"path"
	"server/code/sshcli"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	"gitlab.jkservice.org/smartagent/bdata"
)

func diskLeft(cli *sshcli.Client, dir string) uint64 {
	if dir == "/" {
		return 0
	}
	if _, err := cli.Stat(dir); os.IsNotExist(err) {
		return diskLeft(cli, path.Dir(dir))
	}
	stat, err := cli.StatVFS(dir)
	if err != nil {
		return diskLeft(cli, path.Dir(dir))
	}
	return stat.Bavail * stat.Bsize
}

func memLeft(cli *sshcli.Client) uint64 {
	data, err := cli.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) != 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])
		value = strings.Replace(value, " kB", "", -1)
		if key == "MemFree" {
			t, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return 0
			}
			return t * 1024
		}
	}
	return 0
}

func runOK(conn *websocket.Conn, id string) {
	var rep bdata.Response
	rep.Id = id
	rep.Ok = true
	data, err := proto.Marshal(&rep)
	if err == nil {
		conn.WriteMessage(websocket.BinaryMessage, data)
	}
}

func runErr(conn *websocket.Conn, id, msg string) {
	var rep bdata.Response
	rep.Id = id
	rep.Ok = false
	rep.Msg = msg
	data, err := proto.Marshal(&rep)
	if err == nil {
		conn.WriteMessage(websocket.BinaryMessage, data)
	}
}
