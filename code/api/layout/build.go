package layout

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

func (r *runner) checkAndBuild(dir string, handlers map[string]taskHandler) error {
	var main *Main
	var err error
	if _, err := os.Stat(filepath.Join(dir, "main.yaml")); !os.IsNotExist(err) {
		main, err = r.decodeYaml(dir)
	} else if _, err := os.Stat(filepath.Join(dir, "main.json")); !os.IsNotExist(err) {
		main, err = r.decodeJson(dir)
	} else {
		return errors.New("missing main.yaml or main.json file")
	}
	if err != nil {
		return err
	}

	if len(main.Name) == 0 {
		return errors.New("missing layout name")
	}
	if main.Timeout == 0 {
		main.Timeout = 3600
	}

	for i, t := range main.Tasks {
		if len(t.Name) == 0 {
			return fmt.Errorf("missing task name for task index %d", i+1)
		}
		if len(t.Plugin) == 0 {
			return fmt.Errorf("missing plugin for task [%s]", t.Name)
		}
		handler, ok := handlers[t.Plugin]
		if !ok {
			return fmt.Errorf("plugin %s not supported", t.Plugin)
		}
		err = handler.Check(t)
		if err != nil {
			return err
		}
		var info taskInfo
		info.parent = r
		info.name = t.Name
		info.auth = t.Auth
		info.timeout = time.Duration(t.Timeout) * time.Second
		if len(info.auth) > 0 {
			switch info.auth {
			case "sudo":
			case "su":
			default:
				return fmt.Errorf("unexpected auth argument for task [%s]", t.Name)
			}
			if len(r.user) == 0 {
				return fmt.Errorf("specify auth argument for task [%s] but missing user argument on calling", t.Name)
			}
		}
		if len(t.If) > 0 {
			err = info.parseIf(t.If)
			if err != nil {
				return fmt.Errorf("parse if expression failed for task [%s]: %v", t.Name, err)
			}
		}
		r.tasks = append(r.tasks, handler.Clone(t, info))
	}

	r.deadline = time.Now().Add(time.Duration(main.Timeout) * time.Second)

	return nil
}

func (r *runner) decodeYaml(dir string) (*Main, error) {
	f, err := os.Open(filepath.Join(dir, "main.yaml"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var main Main
	err = yaml.NewDecoder(f).Decode(&main)
	if err != nil {
		return nil, fmt.Errorf("decode main.yaml: %v", err)
	}
	return &main, err
}

func (r *runner) decodeJson(dir string) (*Main, error) {
	f, err := os.Open(filepath.Join(dir, "main.json"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var main Main
	err = yaml.NewDecoder(f).Decode(&main)
	if err != nil {
		return nil, fmt.Errorf("decode main.json: %v", err)
	}
	return &main, err
}
