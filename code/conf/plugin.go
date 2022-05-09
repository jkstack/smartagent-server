package conf

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"server/code/utils"
	"sync"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

type PluginInfo struct {
	Name    string
	Version utils.Version
	OS      string
	Arch    string
	Dir     string
	MD5     [md5.Size]byte
	URI     string
}

type PluginMD5List struct {
	sync.RWMutex
	data map[string]PluginInfo // md5 => info
}

func newPluginMD5List() *PluginMD5List {
	return &PluginMD5List{
		data: make(map[string]PluginInfo),
	}
}

func (list *PluginMD5List) update(enc string, info PluginInfo) {
	list.Lock()
	list.data[enc] = info
	list.Unlock()
}

func (list *PluginMD5List) get(ver utils.Version, os, arch string) *PluginInfo {
	if list == nil {
		return nil
	}
	var ret *PluginInfo
	list.RLock()
	for _, p := range list.data {
		if p.Version.Equal(ver) &&
			p.OS == os && p.Arch == arch {
			ret = &p
			break
		}
	}
	list.RUnlock()
	return ret
}

func (list *PluginMD5List) by(md5 string) *PluginInfo {
	if list == nil {
		return nil
	}
	list.RLock()
	defer list.RUnlock()
	if p, ok := list.data[md5]; ok {
		return &p
	}
	return nil
}

type SupportedItem struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	File string `json:"file"`
}

type Manifest struct {
	Name      string          `json:"name"`
	Version   string          `json:"version"`
	Supported []SupportedItem `json:"supported"`
}

func (cfg *Configure) LoadPlugin() {
	files, err := filepath.Glob(path.Join(cfg.PluginDir, "*")) // name
	runtime.Assert(err)
	for _, file := range files {
		cfg.Lock()
		if _, ok := cfg.PluginList[path.Base(file)]; !ok {
			cfg.PluginList[path.Base(file)] = newPluginMD5List()
		}
		cfg.Unlock()
		files, err := filepath.Glob(path.Join(file, "*")) // version
		runtime.Assert(err)
		cfg.Lock()
		delete(cfg.PluginLatest, path.Base(file))
		cfg.Unlock()
		for _, file := range files {
			version := path.Base(file)
			dir := path.Join(file, "manifest.json")
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				continue
			}
			cfg.loadManifest(dir, version)
		}
	}
}

func (cfg *Configure) loadManifest(dir, version string) {
	f, err := os.Open(dir)
	runtime.Assert(err)
	defer f.Close()
	var mf Manifest
	runtime.Assert(json.NewDecoder(f).Decode(&mf))
	ver, err := utils.ParseVersion(version)
	runtime.Assert(err)
	for _, it := range mf.Supported {
		pd := path.Join(path.Dir(dir), it.File)
		md5, err := utils.MD5Checksum(pd)
		runtime.Assert(err)
		enc := fmt.Sprintf("%x", md5)
		cfg.PluginList[mf.Name].update(enc, PluginInfo{
			Name:    mf.Name,
			Version: ver,
			OS:      it.OS,
			Arch:    it.Arch,
			Dir:     pd,
			MD5:     md5,
			URI:     fmt.Sprintf("/file/plugin/%s/%s", mf.Name, enc),
		})
		logging.Info("load plugin %s, os=%s, arch=%s, version=%s, md5=%x",
			mf.Name, it.OS, it.Arch, ver.String(), md5)
		if ver.Greater(cfg.PluginLatest[mf.Name]) {
			cfg.Lock()
			cfg.PluginLatest[mf.Name] = ver
			cfg.Unlock()
		}
	}
}
