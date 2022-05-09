package layout

import (
	"errors"
	"fmt"
)

var errTimeout = errors.New("timeout")

func errClientNotfound(id string) error {
	return fmt.Errorf("client not found: %s", id)
}

func errPluginNotInstalled(name string) error {
	return fmt.Errorf("plugin not installed: %s", name)
}
