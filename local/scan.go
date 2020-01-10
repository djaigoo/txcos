// Package Files Files

package local

import (
    "os"
    "path/filepath"
    "sync"
    
    "github.com/djaigoo/logkit"
    "github.com/djaigoo/txcos/utils"
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
        s := NewFile(path, info.ModTime().UnixNano())
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
            names, err := utils.ReadDirNames(tmpPath)
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
                    if GIgnore.FindDir(name) {
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
                    if GIgnore.FindFile(name) {
                        continue
                    }
                    s := NewFile(name, info.ModTime().UnixNano())
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
func Check(last []File) (crt, mod, del []File) {
    a, b := 0, 0
    crt = make([]File, 0)
    mod = make([]File, 0)
    del = make([]File, 0)
    for a < len(GFileList) && b < len(last) {
        if !FindSieve(GFileList[a].FilePath()) {
            a++
            continue
        }
        if Compare(GFileList[a], last[b]) == 0 {
            ap := GFileList[a]
            info, err := os.Lstat(ap.FilePath())
            if err != nil {
                logkit.Errorf("os lstat %s, error:%s", ap, err)
                a++
                b++
                continue
            }
            if info.ModTime().Unix() > GFileList[a].UpdateTime.Unix() {
                mod = append(mod, GFileList[a])
            }
            a++
            b++
        } else if Compare(GFileList[a], last[b]) > 0 {
            crt = append(crt, last[b])
            b++
        } else {
            del = append(del, GFileList[a])
            a++
        }
    }
    
    if a == len(GFileList) {
        crt = append(crt, last[b:]...)
    } else {
        for _, file := range GFileList[a:] {
            if !FindSieve(file.FilePath()) {
                a++
                continue
            }
            del = append(del, file)
        }
    }
    return
}
