package hm

import (
	"fmt"
	"net/http"
	lapi "server/code/api"
	"server/code/client"
	"time"

	"github.com/jkstack/anet"
	"github.com/lwch/api"
	"github.com/lwch/runtime"
)

func (h *Handler) static(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")

	cli := clients.Get(id)
	if cli == nil {
		ctx.NotFound("client")
	}

	p := h.cfg.GetPlugin("host.monitor", cli.OS(), cli.Arch())
	if p == nil {
		lapi.PluginNotInstalledErr("host.monitor")
	}

	taskID, err := cli.SendHMStatic(p)
	runtime.Assert(err)
	defer cli.ChanClose(taskID)

	var msg *anet.Msg
	select {
	case msg = <-cli.ChanRead(taskID):
	case <-time.After(api.RequestTimeout):
		ctx.Timeout()
	}

	switch {
	case msg.Type == anet.TypeError:
		ctx.ERR(http.StatusServiceUnavailable, msg.ErrorMsg)
		return
	case msg.Type != anet.TypeHMStaticRep:
		ctx.ERR(http.StatusInternalServerError, fmt.Sprintf("invalid message type: %d", msg.Type))
		return
	}

	type core struct {
		Processor int32  `json:"processor"`
		Model     string `json:"model"`
		Core      int32  `json:"core"`
		Cores     int32  `json:"cores"`
		Physical  int32  `json:"physical"`
	}

	type disk struct {
		Name   string   `json:"name"`
		Type   string   `json:"type"`
		Opts   []string `json:"opts"`
		Total  uint64   `json:"total"`
		INodes uint64   `json:"inodes"`
	}

	type intf struct {
		Index   int      `json:"index"`
		Name    string   `json:"name"`
		Mtu     int      `json:"mtu"`
		Flags   []string `json:"flags"`
		Mac     string   `json:"mac"`
		Address []string `json:"addrs"`
	}

	var ret struct {
		Timestamp       int64  `json:"timestamp"`
		HostName        string `json:"host_name"`
		UpTime          int64  `json:"uptime"`
		OSName          string `json:"os_name"`
		Platform        string `json:"platform"`
		PlatformVersion string `json:"platform_version"`
		KernelArch      string `json:"kernel_arch"`
		KernelVersion   string `json:"kernel_version"`
		PhysicalCore    int    `json:"physical_core"`
		LogicalCore     int    `json:"logical_core"`
		Cores           []core `json:"cores"`
		PhysicalMemory  uint64 `json:"physical_memory"`
		SwapMemory      uint64 `json:"swap_memory"`
		Disks           []disk `json:"disks"`
		Interface       []intf `json:"intfs"`
	}

	info := msg.HMStatic
	ret.Timestamp = info.Time.Unix()
	ret.HostName = info.Host.Name
	ret.UpTime = int64(info.Host.UpTime.Seconds())
	ret.OSName = info.OS.Name
	ret.Platform = info.OS.PlatformName
	ret.PlatformVersion = info.OS.PlatformVersion
	ret.KernelArch = info.Kernel.Arch
	ret.KernelVersion = info.Kernel.Version
	ret.PhysicalCore = info.CPU.Physical
	ret.LogicalCore = info.CPU.Logical
	for _, c := range info.CPU.Cores {
		ret.Cores = append(ret.Cores, core{
			Processor: c.Processor,
			Model:     c.Model,
			Core:      c.Core,
			Cores:     c.Cores,
			Physical:  c.Physical,
		})
	}
	ret.PhysicalMemory = info.Memory.Physical
	ret.SwapMemory = info.Memory.Swap
	for _, d := range info.Disks {
		ret.Disks = append(ret.Disks, disk{
			Name:   d.Name,
			Type:   d.FSType,
			Opts:   d.Opts,
			Total:  d.Total,
			INodes: d.INodes,
		})
	}
	for _, i := range info.Interface {
		ret.Interface = append(ret.Interface, intf{
			Index:   i.Index,
			Name:    i.Name,
			Flags:   i.Flags,
			Mac:     i.Mac,
			Address: i.Address,
		})
	}

	ctx.OK(ret)
}
