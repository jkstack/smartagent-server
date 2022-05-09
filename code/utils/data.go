package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func EncodeData(data []byte) string {
	var str string
	if strings.Contains(http.DetectContentType(data), "text/plain") {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		if _, err := w.Write(data); err == nil {
			w.Close()
			str = "$1$" + base64.StdEncoding.EncodeToString(buf.Bytes())
		}
	}
	if len(str) == 0 {
		str = "$0$" + base64.StdEncoding.EncodeToString(data)
	}
	return str
}

func DecodeData(str string) ([]byte, error) {
	switch {
	case strings.HasPrefix(str, "$0$"):
		str := strings.TrimPrefix(str, "$0$")
		return base64.StdEncoding.DecodeString(str)
	case strings.HasPrefix(str, "$1$"):
		str := strings.TrimPrefix(str, "$1$")
		b64 := base64.NewDecoder(base64.StdEncoding, strings.NewReader(str))
		r, err := gzip.NewReader(b64)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	default:
		return nil, fmt.Errorf("invalid data: %s", str)
	}
}
