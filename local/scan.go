// Package Files Files

package local

import (
    "os"
    "path/filepath"
    "strings"
    "sync"
    
    "github.com/djaigoo/logkit"
    "github.com/djaigoo/txcos/confs"
)

// ScanFile
type ScanFile struct {
    Files []File
}

// NewScanFile new ScanFile
func NewScanFile() *ScanFile {
    return &ScanFile{
        Files: make([]File, 0, 64),
    }
}

func (sf *ScanFile) Walk(path string) (err error) {
    info, err := os.Lstat(path)
    if err != nil {
        return err
    }
    if !info.IsDir() {
        s := NewFile(path)
        sf.Files = append(sf.Files, s)
        return nil
    }
    return sf.walk(path)
}

func (sf *ScanFile) walk(path string) (err error) {
    clen := 1024
    ch := make(chan string, clen)
    ch <- path
    wg := new(sync.WaitGroup)
    once := new(sync.Once)
    wg.Add(1)
    go func() {
        for tmpPath := range ch {
            names, err := confs.ReadDirNames(tmpPath)
            if err != nil {
                logkit.Errorf("read dir %s names %s", tmpPath, err.Error())
                break
            }
            for _, name := range names {
                name = filepath.Join(tmpPath, name)
                info, err := os.Lstat(name)
                if err != nil {
                    logkit.Errorf("file %s error %s", name, err.Error())
                    continue
                }
                if info.IsDir() {
                    if Ignore().FindDir(name) {
                        continue
                    }
                    if len(ch) < clen>>1 {
                        ch <- name
                    } else {
                        wg.Add(1)
                        go func(name string) {
                            defer wg.Done()
                            ch <- name
                        }(name)
                    }
                } else {
                    if Ignore().FindFile(name) {
                        continue
                    }
                    s := NewFile(name)
                    sf.Files = append(sf.Files, s)
                }
            }
            if len(ch) == 0 {
                once.Do(wg.Done)
            }
        }
    }()
    wg.Wait()
    close(ch)
    return nil
}

// Check 查出新建文件，修改文件，删除文件，last必须是sort后的相对路径序列，返回路径为绝对路径
// 会与上一次上传的列表进行对比，所以last必须传入的是本次修改的全部内容
// area 表示扫描核对范围
func Check(area []string, last []File) (crt, mod, del []File) {
    a, b := 0, 0
    crt = make([]File, 0)
    mod = make([]File, 0)
    del = make([]File, 0)
    for a < len(Record().Files) && b < len(last) {
        var e bool
        for _, s := range area {
            if strings.HasPrefix(Record().Files[a].Dir, s) {
                e = true
                break
            }
        }
        if !e || !Sieve().Find(Record().Files[a].FilePath()) {
            a++
            continue
        }
        switch Compare(Record().Files[a], last[b]) {
        case 0:
            ap := Record().Files[a]
            info, err := os.Lstat(ap.FilePath())
            if err != nil {
                logkit.Errorf("os lstat %s, error:%s", ap, err)
                a++
                b++
                continue
            }
            if info.ModTime().UnixNano() > Record().Files[a].UpdateTime {
                mod = append(mod, Record().Files[a])
            }
            a++
            b++
        case 1:
            crt = append(crt, last[b])
            b++
        case -1:
            del = append(del, Record().Files[a])
            a++
        }
    }
    
    if a == len(Record().Files) {
        crt = append(crt, last[b:]...)
    } else {
        for _, file := range Record().Files[a:] {
            if !Sieve().Find(file.FilePath()) {
                a++
                continue
            }
            del = append(del, file)
        }
    }
    return
}
