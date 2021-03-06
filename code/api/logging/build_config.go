package logging

import (
	"sort"

	"github.com/lwch/api"
)

type k8sConfig struct {
	Namespace string   `json:"ns"`
	Names     []string `json:"name"`
	Dir       string   `json:"dir"`
	Api       string   `json:"api"`
	Token     string   `json:"token"`
}

func (cfg *k8sConfig) build(ctx *api.Context) error {
	cfg.Namespace = ctx.XStr("ns")
	cfg.Names = ctx.XCsv("names")
	cfg.Dir = ctx.OStr("dir", "")
	cfg.Api = ctx.XStr("api")
	cfg.Token = ctx.XStr("token")
	sort.Strings(cfg.Names)
	return nil
}

type fileConfig struct {
	Dir string `json:"dir"`
}

func (cfg *fileConfig) build(ctx *api.Context) error {
	cfg.Dir = ctx.OStr("dir", "")
	return nil
}

type dockerConfig struct {
	ContainerName string `json:"ct_name"`
	ContainerTag  string `json:"ct_tag"`
	Dir           string `json:"dir"`
}

func (cfg *dockerConfig) build(ctx *api.Context) error {
	cfg.ContainerName = ctx.XStr("ct_name")
	cfg.ContainerTag = ctx.OStr("ct_tag", "")
	cfg.Dir = ctx.OStr("dir", "")
	return nil
}
