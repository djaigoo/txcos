// Package Files Files

package confs

import (
    "os"
    "path/filepath"
    "sort"
)

const (
    SYS_DIR       = ".cos"
    SYS_CONF      = "conf.toml"
    SYS_YAML_CONF = "conf.yaml"
    SYS_RECORD    = ".record"
    SYS_IGNORE    = ".ignore"
)

var (
    rootPath = ""
)

func SysDir() string {
    return filepath.Join(rootPath, SYS_DIR)
}

func SysConf() string {
    return filepath.Join(rootPath, SYS_DIR, SYS_CONF)
}

func SysYamlConf() string {
    return filepath.Join(rootPath, SYS_DIR, SYS_YAML_CONF)
}

func SysRecord() string {
    return filepath.Join(rootPath, SYS_DIR, SYS_RECORD)
}

func SysIgnore() string {
    return filepath.Join(rootPath, SYS_DIR, SYS_IGNORE)
}

// InitRootPath 将当前路径设置为根路径
func InitRootPath() {
    curPath, err := filepath.Abs(".")
    if err != nil {
        panic(err.Error())
    }
    rootPath = curPath
}

// RootPath 查找当前绝对路径中是否包含.cos文件夹
func RootPath() string {
    if rootPath == "" {
        rootPath = findRoot(".")
    }
    if rootPath == "/" {
        rootPath = ""
    }
    return rootPath
}

func findRoot(path string) string {
    path, _ = filepath.Abs(path)
    ppath := path
    for {
        info, err := os.Lstat(filepath.Join(path, SYS_DIR))
        if err == nil && info.IsDir() {
            return path
        }
        path = filepath.Dir(path)
        if path == ppath {
            break
        }
        ppath = path
    }
    return path
}

// ReadDirNames 获取文件夹中文件 按文件名排序
func ReadDirNames(dirname string) ([]string, error) {
    f, err := os.Open(dirname)
    if err != nil {
        return nil, err
    }
    names, err := f.Readdirnames(-1)
    f.Close()
    if err != nil {
        return nil, err
    }
    sort.Strings(names)
    return names, nil
}

// GetFileName 以根目录为基础返回绝对路径
func GetFileName(path string) (string, error) {
    if !filepath.IsAbs(path) {
        path = filepath.Join(RootPath(), path)
    }
    _, err := os.Lstat(path)
    return path, err
}
