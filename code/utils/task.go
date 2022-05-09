package utils

import (
	"fmt"
	"sync"
	"time"

	"github.com/lwch/runtime"
)

var taskDate string
var taskCnt uint32
var taskLock sync.Mutex

func TaskID() (string, error) {
	now := time.Now()
	taskLock.Lock()
	defer taskLock.Unlock()
	if now.Format("20060102") != taskDate {
		taskDate = now.Format("20060102")
		taskCnt = 0
	}
	rand, err := runtime.UUID(16, "0123456789abcdef")
	if err != nil {
		return "", err
	}
	taskCnt++
	return fmt.Sprintf("%s-%05d-%s", now.Format("20060102"), taskCnt%99999, rand), nil
}
