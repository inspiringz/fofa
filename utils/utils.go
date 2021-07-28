package utils

import (
	"bufio"
	"fofa/logger"
	"os"
)

func ScanFile(filePath string) (querys []string) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		logger.Warn(err.Error())
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		querys = append(querys, sc.Text())
	}
	if err := sc.Err(); err != nil {
		logger.Warn(err.Error())
		return
	}
	return
}
