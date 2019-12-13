// Package Files Files

package files

import (
    "bytes"
    "io/ioutil"
    "path/filepath"
    "strings"
    
    "github.com/djaigoo/txcos/utils"
)

const (
    endMark = '/'
    endLine = '\n'
)

// Ignore
type Ignore struct {
    Files []string
    Dirs  []string
}

// NewIgnore new Ignore
func NewIgnore(path string) *Ignore {
    ret := &Ignore{
        Files: nil,
        Dirs:  nil,
    }
    content, _ := ioutil.ReadFile(path)
    if len(content) == 0 {
        return ret
    }
    lines := bytes.Split(content, []byte{endLine})
    for _, line := range lines {
        if len(line) == 0 {
            continue
        }
        if line[len(line)-1] == endMark {
            ret.Dirs = append(ret.Dirs, string(line))
        } else {
            ret.Files = append(ret.Files, string(line))
        }
    }
    ret.Dirs = append(ret.Dirs, utils.SYS_DIR+string(endMark))
    return ret
}

// Close Close
func (ig *Ignore) Close() error {
    content := bytes.NewBuffer(nil)
    content.WriteString(strings.Join(ig.Dirs, string(endLine)))
    content.Write([]byte{endLine})
    content.WriteString(strings.Join(ig.Files, string(endLine)))
    return ioutil.WriteFile(utils.SysIgnore(), content.Bytes(), 0644)
}

// FindDir
func (ig *Ignore) FindDir(dir string) bool {
    if len(dir) == 0 {
        return false
    }
    _, dir = filepath.Split(dir)
    
    if dir[len(dir)-1] != endMark {
        dir = dir + string(endMark)
    }
    res := false
    for _, d := range ig.Dirs {
        if d == dir {
            res = true
            break
        }
    }
    return res
}

// FindFile FindFile
func (ig *Ignore) FindFile(file string) bool {
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

var GIgnore *Ignore

func InitIgnore() {
    GIgnore = NewIgnore(utils.SysIgnore())
}

func CloseIgnore() (err error) {
    return GIgnore.Close()
}
