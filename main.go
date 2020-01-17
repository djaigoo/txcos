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
    start := time.Now()
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

// pathUniq 去重传入路径参数
// 返回值会打乱传入参数的顺序
func pathUniq(paths []string) []string {
    m := make(map[string]struct{})
    for _, path := range paths {
        m[path] = struct{}{}
    }
    ret := make([]string, 0, len(m))
    for k := range m {
        ret = append(ret, k)
    }
    return ret
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

// check 获取指定目录下有变动的文件
func check(paths ...string) (crt, mod, del []local.File) {
    checkPath := make([]local.File, 0)
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
        checkPath = append(checkPath, sf.Files...)
    }
    checkPath = local.SortAndUniq(checkPath)
    return local.Check(checkPath)
}

// status 查看文档更改状态
func status() error {
    paths := pathUniq(os.Args[2:])
    crt, mod, del := check(paths...)
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

func bucketExec(ctx context.Context, files []local.File, f func(ctx context.Context, file local.File) error) (ret []local.File) {
    bucket := utils.NewTokenBucket(100)
    bucket.Get()
    go func() {
        defer bucket.Put()
        for _, file := range files {
            bucket.Get()
            go func(file local.File) {
                defer bucket.Put()
                err := f(ctx, file)
                if err != nil {
                    logkit.Errorf(err.Error())
                    return
                }
                // TODO 并发不安全
                ret = append(ret, file)
            }(file)
        }
    }()
    bucket.Wait()
    return ret
}

func putRemoteFile(ctx context.Context, file local.File) error {
    data, err := ioutil.ReadFile(file.FilePath())
    if err != nil {
        return errors.Wrapf(err, "read file %s", file.FilePath())
    }
    name := getRemoteName(file)
    if ok, _ := diffRemote(ctx, name, data); ok {
        logkit.Alertf("[push] file %s --> %s no modification", file.FilePath(), name)
        return nil
    }
    err = remote.GClient.Put(ctx, name, bytes.NewBuffer(data))
    if err != nil {
        return errors.Wrapf(err, "[push] cos put %s --> %s", file.FilePath(), name)
    }
    return nil
}

func modRemoteFile(ctx context.Context, file local.File) error {
    logkit.Debugf("%s", file.FilePath())
    data, err := ioutil.ReadFile(file.FilePath())
    if err != nil {
        return errors.Wrapf(err, "read file %s", file.FilePath())
    }
    name := getRemoteName(file)
    if ok, _ := diffRemote(ctx, name, data); ok {
        logkit.Alertf("file %s --> %s no modification", file.FilePath(), name)
        return nil
    }
    err = remote.GClient.Put(ctx, name, bytes.NewBuffer(data))
    if err != nil {
        return errors.Wrapf(err, "cos put %s --> %s", file.FilePath(), name)
    }
    return nil
}

func delRemoteFile(ctx context.Context, file local.File) error {
    name := getRemoteName(file)
    err := remote.GClient.Delete(ctx, name)
    if err != nil {
        return errors.Wrapf(err, "[push] delete %s", file.FilePath())
    }
    return nil
}

func getRemoteFile(ctx context.Context, file local.File) error {
    data, err := ioutil.ReadFile(file.FilePath())
    if err != nil {
        return errors.Wrapf(err, "read file %s", file.FilePath())
    }
    name := getRemoteName(file)
    if ok, _ := diffRemote(ctx, name, data); ok {
        logkit.Alertf("file %s --> %s no modification", file.FilePath(), name)
        return nil
    }
    msg, err := remote.GClient.Get(ctx, name)
    if err != nil {
        return errors.Wrapf(err, "cos put %s --> %s", file.FilePath(), name)
    }
    err = ioutil.WriteFile(file.FilePath(), msg, 0644)
    if err != nil {
        return errors.Wrapf(err, "write file %s --> %s", file.FilePath(), name)
    }
    return nil
}

// push 上传本地修改至cos
func push() error {
    if len(os.Args) == 2 {
        os.Args = append(os.Args, ".")
    }
    paths := pathUniq(os.Args[2:])
    crt, mod, del := check(paths...)
    ctx := context.Background()
    tcrt := bucketExec(ctx, crt, putRemoteFile)
    tmod := bucketExec(ctx, mod, modRemoteFile)
    tdel := bucketExec(ctx, del, delRemoteFile)
    for _, file := range tcrt {
        logkit.Infof("[push] new file %s --> %s succeed", file.FilePath(), getRemoteName(file))
    }
    
    for _, file := range tmod {
        logkit.Infof("[push] mod file %s --> %s succeed", file.FilePath(), getRemoteName(file))
    }
    
    for _, file := range tdel {
        logkit.Infof("[push] del file %s --> %s succeed", file.FilePath(), getRemoteName(file))
    }
    
    local.Merge(&local.GFileList, tcrt, tmod, tdel)
    return local.CloseRecord()
}

// pull 拉取远端文件
func pull() error {
    if len(os.Args) == 2 {
        os.Args = append(os.Args, ".")
    }
    paths := pathUniq(os.Args[2:])
    _, mod, _ := check(paths...)
    ctx := context.Background()
    tmod := bucketExec(ctx, mod, getRemoteFile)
    for _, file := range tmod {
        logkit.Infof("[pull] modify %s --> %s succeed", file.FilePath(), getRemoteName(file))
    }
    
    local.Merge(&local.GFileList, nil, tmod, nil)
    return local.CloseRecord()
}

// diffRemote 对比本地与远端文件内容是否相同
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
