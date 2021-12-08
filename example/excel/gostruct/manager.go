// Code generated converter.
// DO NOT EDIT!
package gostruct

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ITable interface {
	LoadJson(string) error
	LoadXlsx(string) error
}

var ManagerSingle = &Manager{
	Tables: make(map[string]*Table),
}

type Table struct {
	T   ITable
	MD5 string
}

type Manager struct {
	Path   string
	Tables map[string]*Table
}

func (m *Manager) Register(name string, t ITable) {
	m.Tables[name] = &Table{T: t}
}

func (m *Manager) Load(name string) error {
	t, ok := m.Tables[name]
	if !ok {
		return nil
	}

	p := filepath.Join(m.Path, name)
	if 92 != os.PathSeparator { // 兼容其他平台的路径分隔符
		p = strings.ReplaceAll(p, string(92), string(os.PathSeparator))
	}

	pXlsx := fmt.Sprintf("%s.xlsx", p)
	md5str, err := getMD5(pXlsx)
	if err == nil {
		if md5str != t.MD5 {
			if err = t.T.LoadXlsx(pXlsx); err != nil {
				return err
			}
			t.MD5 = md5str
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	pJson := fmt.Sprintf("%s.json", p)
	md5str, err = getMD5(pJson)
	if err == nil {
		if md5str != t.MD5 {
			if err = t.T.LoadJson(pJson); err != nil {
				return err
			}
			t.MD5 = md5str
		}
		return nil
	}
	if os.IsNotExist(err) {
		return errors.New("file no exist")
	}
	return err
}

func getMD5(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func init() {
	ManagerSingle.Register("Test", TestSingle)
}

// Init 设置文件所在目录，并加载数据
func Init(path string) error {
	ManagerSingle.Path = path
	ch := make(chan error, len(ManagerSingle.Tables))
	g := sync.WaitGroup{}
	for name := range ManagerSingle.Tables {
		g.Add(1)
		go func(name string) {
			defer g.Done()
			if err := ManagerSingle.Load(name); err != nil {
				ch <- errors.New(fmt.Sprintf("manager: load file error:%v path:%s filename:%v", err, path, name))
			}
		}(name)
	}
	g.Wait()
	close(ch)
	return <-ch
}

// Load 重新加载文件数据
// name 文件名
func Load(name string) error {
	return ManagerSingle.Load(strings.TrimRight(filepath.Base(name), filepath.Ext(name)))
}
