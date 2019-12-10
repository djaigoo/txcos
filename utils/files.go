// Package Files Files

package utils

import (
    "os"
    "path/filepath"
)

var (
    ROOT_PATH  = ""
    SYS_DIR    = ".cos"
    SYS_CONF   = "conf.toml"
    SYS_RECORD = ".record"
    SYS_IGNORE = ".ignore"
)

func RootPath() string {
    return ROOT_PATH
}

func SysDir() string {
    return filepath.Join(ROOT_PATH, SYS_DIR)
}

func SysConf() string {
    return filepath.Join(ROOT_PATH, SYS_DIR, SYS_CONF)
}

func SysRecord() string {
    return filepath.Join(ROOT_PATH, SYS_DIR, SYS_RECORD)
}

func SysIgnore() string {
    return filepath.Join(ROOT_PATH, SYS_DIR, SYS_IGNORE)
}

func FindRoot() string {
    path := findRoot(".")
    return path
}

func findRoot(path string) string {
    path, _ = filepath.Abs(path)
    for {
        ppath := filepath.Join(path, SYS_DIR)
        info, err := os.Lstat(ppath)
        if err == nil && info.IsDir() {
            return path
        }
        // if err != nil || !info.IsDir() {
        //     panic(err.Error())
        //     continue
        // }
        spath := path
        path = filepath.Dir(path)
        if path == spath {
            break
        }
    }
    return path
}

// ReadDirNames 获取文件夹中文件
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
    // sort.Strings(names)
    return names, nil
}
