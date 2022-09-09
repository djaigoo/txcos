// txcos 交互式快速上传本地修改文件至cos
package main

import (
    "bytes"
    "context"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "time"
    
    "github.com/djaigoo/logkit"
    "github.com/djaigoo/txcos/confs"
    "github.com/djaigoo/txcos/local"
    "github.com/djaigoo/txcos/remote"
    "github.com/djaigoo/txcos/utils"
    "github.com/djaigoo/txcos/xerror"
    "github.com/pkg/errors"
)

func init() {
    register("-i, init", "initialize related configuration items", initialize)
    register("-h, help", "show usage", help)
    register("-l, pull", "pull remote file", pull)
    register("-p, push", "push add information", push)
    register("-s, status", "get status", status)
    register("-ci, cimage", "from clipboard get image", clipboardImage)
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println(usage())
        return
    }
    start := time.Now()
    showtime := false
    defer func() {
        if showtime {
            logkit.Debugf("exec success cost %s", time.Now().Sub(start).String())
        }
    }()
    
    cmd := os.Args[1]
    err := do(cmd)
    if err != nil && err == xerror.ErrCmdNotExist {
        fmt.Println("not found command ", cmd)
        fmt.Println(usage())
        return
    }
    showtime = true
    if err != nil {
        logkit.Errorf("exec error %s", err)
        return
    }
}

// initialize 初始化txcos系统
func initialize() (err error) {
    confs.InitRootPath()
    _, err = os.Lstat(confs.SysDir())
    if err == nil {
        logkit.Infof("%s already exists in the current directory", confs.SYS_DIR)
        return
    } else if e, ok := err.(*os.PathError); ok {
        if e.Err.Error() != "no such file or directory" {
            logkit.Errorf("%s", e.Err.Error())
            return
        }
    }
    err = os.Mkdir(confs.SysDir(), 0755)
    if err != nil {
        return errors.Wrap(err, "make dir")
    }
    err = ioutil.WriteFile(confs.SysConf(), confs.DefaultConf, 0644)
    if err != nil {
        return errors.Wrap(err, "write sys conf")
    }
    err = local.Record().Close()
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
// paths 绝对路径
func check(paths ...string) (crt, mod, del []local.File) {
    checkPath := make([]local.File, 0)
    for _, p := range paths {
        sf := local.NewScanFile()
        err := sf.Walk(p)
        if err != nil {
            logkit.Errorf("walk error %s", err.Error())
            continue
        }
        checkPath = append(checkPath, sf.Files...)
    }
    local.Sort(checkPath)
    checkPath = SortedUniq(checkPath)
    return local.Check(paths, checkPath)
}

// SortedUniq 有序列表去重
func SortedUniq[T string | local.File](paths []T) []T {
    s, f := 0, 0
    for f < len(paths) {
        if paths[f] == paths[s] {
            f++
            continue
        }
        s++
        paths[s] = paths[f]
        f++
    }
    return paths[:s+1]
}

// pathUniq 去重传入路径参数
// 返回值会排序
func pathUniq(paths []string) []string {
    sort.Strings(paths)
    return SortedUniq(paths)
}

// getOperDirs 获取操作目录 返回绝对路径
func getOperDirs() []string {
    ps := os.Args[2:]
    if len(ps) == 0 {
        ps = local.Sieve().Paths()
    } else {
        for i, p := range ps {
            pp, err := filepath.Abs(filepath.Join(".", p))
            if err != nil {
                panic("invalid path: " + p)
            }
            ps[i] = pp
        }
    }
    // 去重排序
    ps = pathUniq(ps)
    return ps
}

// status 查看文档更改状态
func status() error {
    ps := getOperDirs()
    if len(ps) == 0 {
        return nil
    }
    crt, mod, del := check(ps...)
    for _, file := range crt {
        logkit.Infof("create file %s", strings.TrimPrefix(file.FilePath(), confs.RootPath()))
    }
    for _, file := range mod {
        logkit.Alertf("modify file %s", strings.TrimPrefix(file.FilePath(), confs.RootPath()))
    }
    for _, file := range del {
        logkit.Warnf("delete file %s", strings.TrimPrefix(file.FilePath(), confs.RootPath()))
    }
    return nil
}

// getRemoteName 获取远端文件名
func getRemoteName(f local.File) string {
    return confs.CosPathMap(f.FilePath())
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
    err = remote.CosClient().Put(ctx, name, bytes.NewBuffer(data))
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
    err = remote.CosClient().Put(ctx, name, bytes.NewBuffer(data))
    if err != nil {
        return errors.Wrapf(err, "cos put %s --> %s", file.FilePath(), name)
    }
    return nil
}

func delRemoteFile(ctx context.Context, file local.File) error {
    name := getRemoteName(file)
    err := remote.CosClient().Delete(ctx, name)
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
    msg, err := remote.CosClient().Get(ctx, name)
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
    paths := getOperDirs()
    if len(paths) == 0 {
        return nil
    }
    crt, mod, del := check(paths...)
    ctx := context.Background()
    tcrt := bucketExec(ctx, crt, putRemoteFile)
    tmod := bucketExec(ctx, mod, modRemoteFile)
    tdel := bucketExec(ctx, del, delRemoteFile)
    for _, file := range tcrt {
        logkit.Infof("[push] create file %s --> %s succeed", file.FilePath(), confs.CosPathAbs(getRemoteName(file)))
    }
    
    for _, file := range tmod {
        logkit.Alertf("[push] modify file %s --> %s succeed", file.FilePath(), confs.CosPathAbs(getRemoteName(file)))
    }
    
    for _, file := range tdel {
        logkit.Warnf("[push] delete file %s --> %s succeed", file.FilePath(), confs.CosPathAbs(getRemoteName(file)))
    }
    
    local.Merge(&local.Record().Files, tcrt, tmod, tdel)
    return local.Record().Close()
}

// pull 拉取远端文件
func pull() error {
    paths := getOperDirs()
    if len(paths) == 0 {
        return nil
    }
    _, mod, _ := check(paths...)
    ctx := context.Background()
    tmod := bucketExec(ctx, mod, getRemoteFile)
    for _, file := range tmod {
        logkit.Infof("[pull] modify %s --> %s succeed", file.FilePath(), getRemoteName(file))
    }
    
    local.Merge(&local.Record().Files, nil, tmod, nil)
    return local.Record().Close()
}

// diffRemote 对比本地与远端文件内容是否相同
func diffRemote(ctx context.Context, remoteName string, content []byte) (bool, error) {
    lmd5 := utils.GetMD5(content)
    rmd5, err := remote.CosClient().GetFileMD5(ctx, remoteName)
    if err != nil {
        return false, errors.Wrap(err, "get file md5")
    }
    if strings.Compare(lmd5, rmd5) == 0 {
        return true, nil
    }
    return false, nil
}

// clipboardImage 将剪贴板图片上传
func clipboardImage() error {
    name, data, err := utils.GetClipboardImage()
    if err != nil {
        return err
    }
    err = remote.CosClient().Put(context.Background(), confs.CosPathClipboard(name), bytes.NewBuffer(data))
    if err != nil {
        return err
    }
    name = confs.CosPathAbsClipboard(name)
    fmt.Println(name)
    utils.SetClipboardText([]byte(name))
    return nil
}
