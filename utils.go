package sysutils

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

var (
	separator               string
	TimeLocationShanghai, _ = time.LoadLocation("Asia/Shanghai")
)

func init() {
	if os.IsPathSeparator('\\') { //前边的判断是否是系统的分隔符
		separator = "\\"
	} else {
		separator = "/"
	}
}

// PathSeparator 返回系统路径分隔符
func PathSeparator() string {
	return separator
}

// PathJoin 判断系统路径分隔符并返回对应的路径字符串拼接结果
func PathJoin(path ...string) string {
	return strings.Join(path, separator)
}

// GetCurPath 获取当前文件执行的路径
func GetCurPath() string {
	file, _ := exec.LookPath(os.Args[0])

	path, _ := filepath.Abs(file)

	rst := filepath.Dir(path)

	return rst
}

// ObjectToURLValues 把json结构体转换为url.values
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

type Time struct {
	time.Time
}

func (t Time) MarshalText() ([]byte, error) {
	return []byte(t.Format("2006-01-02 15:04:05")), nil
}

func (t *Time) UnmarshalText(str []byte) error {
	tt, err := time.Parse("2006-01-02 15:04:05", string(str))
	if err != nil {
		return err
	}
	*t = Time{tt}
	return nil
}

func (t Time) String() string {
	return t.Format("2006-01-02 15:04:05")
}

func Now() Time {
	return Time{time.Now().Local().In(TimeLocationShanghai)}
}

// 字节的单位转换 保留两位小数
func FormatSize(fileSize int64) (size string) {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}

type Expireable interface {
	Key() string
	IsExpired() bool
}

type ExpireableStorage struct {
	data    map[string]Expireable
	crontab *cron.Cron
}

func (s *ExpireableStorage) Add(delta Expireable) {
	s.data[delta.Key()] = delta
}

func (s *ExpireableStorage) Del(delta Expireable) {
	delete(s.data, delta.Key())
}

func (s *ExpireableStorage) Get(key string) (Expireable, error) {
	if v, ok := s.data[key]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("invalid key")
}

func (s *ExpireableStorage) Destroy() {
	for k := range s.data {
		delete(s.data, k)
	}
	ctx := s.crontab.Stop()
	<-ctx.Done()
	s = nil
}

func (s *ExpireableStorage) Run() {
	count := 0
	for k, v := range s.data {
		if v.IsExpired() {
			delete(s.data, k)
			count++
		}
	}
}

func NewExpireableStorage(spec string) (*ExpireableStorage, error) {
	es := &ExpireableStorage{
		data:    make(map[string]Expireable),
		crontab: cron.New(cron.WithLocation(TimeLocationShanghai)),
	}
	_, err := cron.ParseStandard(spec)
	if err != nil {
		return nil, err
	}
	_, err = es.crontab.AddJob(spec, es)
	if err != nil {
		return nil, err
	}
	es.crontab.Start()
	return es, nil
}
