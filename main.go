// txcos 交互式快速上传本地修改文件至cos
package main

import (
    "bytes"
    "context"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/djaigoo/logkit"
    "github.com/djaigoo/txcos/confs"
    "github.com/djaigoo/txcos/local"
    "github.com/djaigoo/txcos/remote"
    "github.com/djaigoo/txcos/utils"
    "github.com/pkg/errors"
)

func init() {
    register("-i, init", "initialize related configuration items", initialize)
    register("-h, help", "show usage", help)
    register("-l, pull", "pull remote file", pull)
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
        local.InitIgnore()
        local.InitRecord()
        if len(os.Args) == 2 {
            path, err := utils.GetFileName(confs.GCos.DefaultPath)
            if err != nil {
                logkit.Errorf("invalid default path %s", err.Error())
                return
            }
            os.Args = append(os.Args, path)
        }
        local.InitSieve(os.Args[2:]...)
        remote.Init()
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

// close 优雅的关闭
func close() (err error) {
    // err = files.CloseIgnore()
    return
}

// initialize 初始化txcos系统
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

// help 打印帮助文档
func help() error {
    fmt.Println(usage())
    return nil
}

func check(paths ...string) (crt, mod, del []local.File) {
    for _, p := range paths {
        path, err := filepath.Abs(p)
        if err != nil {
            logkit.Errorf("invalid path %s", p)
            continue
        }
        sf := local.NewScanFile()
        err = sf.Walk(path)
        if err != nil {
            logkit.Errorf("walk error %s", err.Error())
            continue
        }
        local.Sort(sf.Files)
        tcrt, tmod, tdel := local.Check(sf.Files)
        crt = append(crt, tcrt...)
        mod = append(mod, tmod...)
        del = append(del, tdel...)
    }
    return
}

// status 查看文档更改状态
func status() error {
    crt, mod, del := check(os.Args[2:]...)
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

// getRemoteName 获取远端文件名
func getRemoteName(f local.File) string {
    dir := confs.PathMap(f.Dir)
    path := strings.TrimPrefix(dir, utils.RootPath())
    return filepath.Join(path, f.Name)
}

// push 上传本地修改至cos
func push() error {
    if len(os.Args) == 2 {
        os.Args = append(os.Args, ".")
    }
    crt, mod, del := check(os.Args[2:]...)
    tcrt := make([]local.File, 0, len(crt))
    tmod := make([]local.File, 0, len(mod))
    tdel := make([]local.File, 0, len(del))
    ctx := context.Background()
    bucket := utils.NewTokenBucket(100)
    bucket.Get()
    go func() {
        defer bucket.Put()
        for _, file := range crt {
            bucket.Get()
            go func(file local.File) {
                defer bucket.Put()
                data, err := ioutil.ReadFile(file.FilePath())
                if err != nil {
                    logkit.Errorf("[push] read file %s error %s", file.FilePath(), err.Error())
                    return
                }
                name := getRemoteName(file)
                err = remote.GClient.Put(ctx, name, bytes.NewBuffer(data))
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
            go func(file local.File) {
                defer bucket.Put()
                data, err := ioutil.ReadFile(file.FilePath())
                if err != nil {
                    logkit.Errorf("[push] read file %s error %s", file.FilePath(), err.Error())
                    return
                }
                name := getRemoteName(file)
                if ok, _ := diffRemote(ctx, name, data); ok {
                    tmod = append(tmod, file)
                    logkit.Alertf("[push] file %s --> %s no modification", file.FilePath(), name)
                    return
                }
                err = remote.GClient.Put(ctx, name, bytes.NewBuffer(data))
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
            go func(file local.File) {
                defer bucket.Put()
                name := getRemoteName(file)
                err := remote.GClient.Delete(ctx, name)
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
    local.Merge(&local.GFileList, tcrt, tmod, tdel)
    return local.CloseRecord()
}

// pull 拉取远端文件
func pull() error {
    if len(os.Args) == 2 {
        os.Args = append(os.Args, ".")
    }
    _, mod, _ := check(os.Args[2:]...)
    tmod := make([]local.File, 0, len(mod))
    ctx := context.Background()
    bucket := utils.NewTokenBucket(100)
    bucket.Get()
    go func() {
        defer bucket.Put()
        for _, file := range mod {
            bucket.Get()
            go func(file local.File) {
                defer bucket.Put()
                data, err := ioutil.ReadFile(file.FilePath())
                if err != nil {
                    logkit.Errorf("[pull] read file %s error %s", file.FilePath(), err.Error())
                    return
                }
                name := getRemoteName(file)
                if ok, _ := diffRemote(ctx, name, data); ok {
                    tmod = append(tmod, file)
                    logkit.Alertf("[pull] file %s --> %s no modification", file.FilePath(), name)
                    return
                }
                msg, err := remote.GClient.Get(ctx, name)
                if err != nil {
                    logkit.Errorf("[pull] cos put %s --> %s error %s", file.FilePath(), name, err.Error())
                    return
                }
                err = ioutil.WriteFile(file.FilePath(), msg, 0644)
                if err != nil {
                    logkit.Errorf("[pull] write file %s --> %s error %s", file.FilePath(), name, err.Error())
                    return
                }
                tmod = append(tmod, file)
                logkit.Infof("[pull] modify %s --> %s succeed", file.FilePath(), name)
            }(file)
        }
    }()
    bucket.Wait()
    local.Merge(&local.GFileList, nil, tmod, nil)
    return local.CloseRecord()
}

func diffRemote(ctx context.Context, remoteName string, content []byte) (bool, error) {
    lmd5 := utils.GetMD5(content)
    rmd5, err := remote.GClient.GetFileMD5(ctx, remoteName)
    if err != nil {
        return false, errors.Wrap(err, "get file md5")
    }
    if strings.Compare(lmd5, rmd5) == 0 {
        return true, nil
    }
    return false, nil
}
