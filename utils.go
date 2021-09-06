package sysutils

import (
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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

//GetCurPath 获取当前文件执行的路径
func GetCurPath() string {
	file, _ := exec.LookPath(os.Args[0])

	path, _ := filepath.Abs(file)

	rst := filepath.Dir(path)

	return rst
}

//ObjectToURLValues 把json结构体转换为url.values
func ObjectToURLValues(p interface{}) url.Values {
	t := reflect.TypeOf(p)
	v := reflect.ValueOf(p)
	params := url.Values{}
	if v.Kind() != reflect.Struct {
		return params
	}
	params = getParams(t, v)
	return params
}
func getParams(t reflect.Type, v reflect.Value) url.Values {
	params := url.Values{}
	if v.Kind() != reflect.Struct {
		return params
	}
	for i := 0; i < v.NumField(); i++ {
		f := t.Field(i)
		if f.Type.Kind() == reflect.Struct {
			tParams := getParams(f.Type, v.Field(i))
			for kk, vv := range tParams {
				params.Set(kk, vv[0])
			}
		} else {
			name, ok := f.Tag.Lookup("form")
			if !ok {
				continue
			}
			vv := v.FieldByName(f.Name).String()
			if vv == "" {
				continue
			}
			params.Set(name, vv)
		}
	}
	return params
}
