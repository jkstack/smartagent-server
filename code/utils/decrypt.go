package utils

import (
	"encoding/base64"
	"strings"

	"github.com/jkstack/anet"
	"github.com/lwch/runtime"
)

func DecryptPass(pass string) string {
	if strings.HasPrefix(pass, "%1%") {
		raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(pass, "%1%"))
		runtime.Assert(err)
		dec, err := anet.Decrypt(raw)
		runtime.Assert(err)
		pass = string(dec)
	}
	return pass
}
