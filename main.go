// txcos 交互式快速上传本地修改文件至cos
package main

import (
    "bytes"
    "context"
    "fmt"
    "github.com/pkg/errors"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/djaigoo/logkit"
    "github.com/djaigoo/txcos/confs"
    "github.com/djaigoo/txcos/local/files"
    "github.com/djaigoo/txcos/remote/cos"
    "github.com/djaigoo/txcos/utils"
)

func init() {
    register("-i, init", "initialize related configuration items", initialize)
    register("-h, help", "show usage", help)
    register("-p, push", "push add information", push)
    register("-s, status", "get status", status)
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println(usage())
        return
    }
    if os.Args[1] != "init" {
        err := confs.InitConf()
        if err != nil {
            logkit.Errorf("%s", err.Error())
            return
        }
        files.InitIgnore()
        files.InitRecord()
        if len(os.Args) == 2 {
            path, err := utils.GetFileName(confs.GCos.DefaultPath)
            if err != nil {
                logkit.Errorf("invalid default path %s", err.Error())
                return
            }
            os.Args = append(os.Args, path)
        }
        files.InitSieve(os.Args[2:]...)
        cos.Init()
    }
    defer close()
    
    start := time.Now()
    cmd := os.Args[1]
    err := do(cmd)
    if err != nil {
        logkit.Errorf("exec error %s", err)
        fmt.Println(usage())
        return
    }
    
    logkit.Infof("exec success cost %s", time.Now().Sub(start).String())
}

func close() (err error) {
    // err = files.CloseIgnore()
    return
}

func initialize() (err error) {
    utils.ROOT_PATH, err = filepath.Abs(".")
    if err != nil {
        return errors.Wrap(err, "get abs path")
    }
    _, err = os.Lstat(utils.SysDir())
    if err == nil {
        logkit.Infof("%s already exists in the current directory", utils.SYS_DIR)
        return
    } else if e, ok := err.(*os.PathError); ok {
        if e.Err.Error() != "no such file or directory" {
            logkit.Errorf("%s", e.Err.Error())
            return
        }
    }
    err = os.Mkdir(utils.SysDir(), 0755)
    if err != nil {
        return errors.Wrap(err, "make dir")
    }
    err = ioutil.WriteFile(utils.SysConf(), []byte(confs.DefaultConf), 0644)
    if err != nil {
        return errors.Wrap(err, "write sys conf")
    }
    err = ioutil.WriteFile(utils.SysIgnore(), nil, 0644)
    if err != nil {
        return errors.Wrap(err, "write sys ignore")
    }
    err = ioutil.WriteFile(utils.SysRecord(), []byte("[]"), 0644)
    if err != nil {
        return errors.Wrap(err, "write sys record")
    }
    return nil
}

func help() error {
    fmt.Println(usage())
    return nil
}

func check(paths ...string) (crt, mod, del []files.File, err error) {
    for _, p := range paths {
        path, err := filepath.Abs(p)
        if err != nil {
            logkit.Errorf("invalid path %s", p)
            continue
        }
        sf := files.NewScanFile()
        err = sf.Walk(path)
        if err != nil {
            logkit.Errorf("walk error %s", err.Error())
            continue
        }
        files.Sort(sf.Files)
        tcrt, tmod, tdel := files.Check(sf.Files)
        crt = append(crt, tcrt...)
        mod = append(mod, tmod...)
        del = append(del, tdel...)
    }
    return
}

func status() error {
    crt, mod, del, err := check(os.Args[2:]...)
    if err != nil {
        return err
    }
    for _, file := range crt {
        logkit.Infof("new file %s", strings.TrimPrefix(file.FilePath(), utils.RootPath()))
    }
    for _, file := range mod {
        logkit.Infof("modify file %s", strings.TrimPrefix(file.FilePath(), utils.RootPath()))
    }
    for _, file := range del {
        logkit.Infof("delete file %s", strings.TrimPrefix(file.FilePath(), utils.RootPath()))
    }
    return nil
}

func getPushName(f files.File) string {
    dir := confs.PathMap(f.Dir)
    path := strings.TrimPrefix(dir, utils.RootPath())
    return filepath.Join(path, f.Name)
}

func push() error {
    if len(os.Args) == 2 {
        os.Args = append(os.Args, ".")
    }
    crt, mod, del, err := check(os.Args[2:]...)
    if err != nil {
        return err
    }
    tcrt := make([]files.File, 0, len(crt))
    tmod := make([]files.File, 0, len(mod))
    tdel := make([]files.File, 0, len(del))
    ctx := context.Background()
    bucket := utils.NewTokenBucket(100)
    bucket.Get()
    go func() {
        defer bucket.Put()
        for _, file := range crt {
            bucket.Get()
            go func(file files.File) {
                defer bucket.Put()
                data, err := ioutil.ReadFile(file.FilePath())
                if err != nil {
                    logkit.Errorf("[push] read file %s error %s", file.FilePath(), err.Error())
                    return
                }
                name := getPushName(file)
                err = cos.GClient.Put(ctx, name, bytes.NewBuffer(data))
                if err != nil {
                    logkit.Errorf("[push] cos put %s --> %s error %s", file.FilePath(), name, err.Error())
                    return
                }
                tcrt = append(tcrt, file)
                logkit.Infof("[push] new file %s --> %s succeed", file.FilePath(), name)
            }(file)
        }
    }()
    
    bucket.Get()
    go func() {
        defer bucket.Put()
        for _, file := range mod {
            bucket.Get()
            go func(file files.File) {
                defer bucket.Put()
                data, err := ioutil.ReadFile(file.FilePath())
                if err != nil {
                    logkit.Errorf("[push] read file %s error %s", file.FilePath(), err.Error())
                    return
                }
                name := getPushName(file)
                err = cos.GClient.Put(ctx, name, bytes.NewBuffer(data))
                if err != nil {
                    logkit.Errorf("[push] cos put %s --> %s error %s", file.FilePath(), name, err.Error())
                    return
                }
                tmod = append(tmod, file)
                logkit.Infof("[push] modify %s --> %s succeed", file.FilePath(), name)
            }(file)
        }
    }()
    
    bucket.Get()
    go func() {
        defer bucket.Put()
        for _, file := range del {
            bucket.Get()
            go func(file files.File) {
                defer bucket.Put()
                name := getPushName(file)
                err := cos.GClient.Delete(ctx, name)
                if err != nil {
                    logkit.Errorf("[push] delete %s error %s", file.FilePath(), err.Error())
                    return
                }
                tdel = append(tdel, file)
                logkit.Infof("[push] delete %s --> %s succeed", file.FilePath(), name)
            }(file)
        }
    }()
    
    bucket.Wait()
    files.Merge(&files.GFileList, tcrt, tmod, tdel)
    return files.CloseRecord()
}
