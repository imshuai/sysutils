package sysutils

import (
	"os"
	"strings"
)

var (
	separator string
)

//PathSeparator 返回系统路径分隔符
func PathSeparator() string {
	return separator
}

//PathJoin 判断系统路径分隔符并返回对应的路径字符串拼接结果
func PathJoin(path ...string) string {
	return strings.Join(path, separator)
}

func init() {
	if os.IsPathSeparator('\\') { //前边的判断是否是系统的分隔符
		separator = "\\"
	} else {
		separator = "/"
	}
}
