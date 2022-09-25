package util

import (
	"os"
)

// 如果目录不存在则创建目录， 如果存在则赋予0777权限
func MkDirIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	return os.MkdirAll(path, os.ModeDir|os.ModePerm)
}

func FileExist(addr string) bool {
	s, err := os.Stat(addr)
	if err != nil {
		return false
	}
	return !s.IsDir()
}
