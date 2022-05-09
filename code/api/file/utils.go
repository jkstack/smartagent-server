package file

import (
	"io"
	"os"
	"server/code/utils"

	"github.com/jkstack/anet"
)

const blockSize = 4096

func fillFile(f *os.File, size uint64) error {
	left := size
	dummy := make([]byte, blockSize)
	for left > 0 {
		if left >= blockSize {
			_, err := f.Write(dummy)
			if err != nil {
				return err
			}
			left -= blockSize
			continue
		}
		dummy = make([]byte, left)
		n, err := f.Write(dummy)
		if err != nil {
			return err
		}
		left -= uint64(n)
	}
	return nil
}

func writeFile(f *os.File, data *anet.DownloadData) (int, error) {
	_, err := f.Seek(int64(data.Offset), io.SeekStart)
	if err != nil {
		return 0, err
	}
	dec, err := utils.DecodeData(data.Data)
	if err != nil {
		return 0, err
	}
	return f.Write(dec)
}
