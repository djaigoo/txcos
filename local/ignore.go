// Package Files Files

package local

import (
    "path/filepath"
    "strings"
    
    "github.com/djaigoo/txcos/confs"
)

// ignoreImp 忽略上传文件/文件夹
type ignoreImp struct {
    Files []string
    Dirs  []string
}

// NewIgnore new ignoreImp
func NewIgnore() *ignoreImp {
    ret := &ignoreImp{
        Files: confs.YamlConf().Ignore.Files,
        Dirs:  confs.YamlConf().Ignore.Dirs,
    }
    ret.Dirs = append(ret.Dirs, confs.SYS_DIR)
    return ret
}

// FindDir
func (ig *ignoreImp) FindDir(dir string) bool {
    if len(dir) == 0 {
        return false
    }
    _, name := filepath.Split(dir)
    strings.TrimSuffix(name, "/")
    res := false
    for _, d := range ig.Dirs {
        if d == name {
            res = true
            break
        }
    }
    return res
}

// FindFile FindFile
func (ig *ignoreImp) FindFile(file string) bool {
    if len(file) == 0 {
        return false
    }
    _, file = filepath.Split(file)
    res := false
    for _, d := range ig.Files {
        if d == file {
            res = true
            break
        }
    }
    return res
}

var ignore *ignoreImp

func Ignore() *ignoreImp {
    if ignore == nil {
        ignore = NewIgnore()
    }
    if ignore == nil {
        panic("ignore nil")
    }
    return ignore
}
