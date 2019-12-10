// Package conf conf

package confs

import (
    "github.com/djaigoo/logkit"
    "io/ioutil"
    "path/filepath"
    "strings"
    
    "github.com/BurntSushi/toml"
    "github.com/djaigoo/txcos/utils"
    "github.com/djaigoo/txcos/xerror"
    "github.com/pkg/errors"
)

// Conf
type Conf struct {
    SecretId  string `toml:"secret_id,omitempty"`
    SecretKey string `toml:"secret_key,omitempty"`
    AppId     string `toml:"app_id,omitempty"`
    Host      string `toml:"host,omitempty"`
    Bucket    string `toml:"bucket,omitempty"`
    Region    string `toml:"region,omitempty"`
    
    // 路由映射，格式"public:/,source:/source"，
    // 前路径表示对于配置文件目录的父目录的相对路径，
    // 后路径表示对于cos存储的路由路径，必须前缀'/'
    MapPath string `toml:"map_path"`
    
    DefaultPath string `toml:"default_path"`
}

// NewConf new Conf
func NewConf() (*Conf, error) {
    utils.ROOT_PATH = utils.FindRoot()
    if utils.ROOT_PATH == "" {
        return nil, xerror.ErrNotFoundRoot
    }
    cnt, err := ioutil.ReadFile(utils.SysConf())
    if err != nil {
        panic(errors.Wrap(err, "read system conf file"))
    }
    ret := new(Conf)
    err = toml.Unmarshal(cnt, ret)
    if err != nil {
        return nil, errors.Wrap(err, "toml unmarshal")
    }
    if ret.DefaultPath == "" {
        ret.DefaultPath = "."
    }
    return ret, nil
}

var GCos *Conf
var pathMap map[string]string

func InitConf() error {
    var err error
    GCos, err = NewConf()
    if err != nil {
        return errors.Wrap(err, "new conf")
    }
    pathMap = make(map[string]string)
    unit := strings.Split(GCos.MapPath, ",")
    for _, u := range unit {
        kv := strings.Split(u, ":")
        if len(kv) != 2 {
            continue
        }
        if len(kv[0]) == 0 || len(kv[1]) == 0 {
            continue
        }
        if !filepath.IsAbs(kv[0]) {
            kv[0], err = filepath.Abs(kv[0])
            if err != nil {
                logkit.Errorf("map_path specifies a local illegal path")
                continue
            }
        }
        if !filepath.IsAbs(kv[1]) {
            kv[1] = filepath.Join(utils.RootPath(), kv[1])
        }
        pathMap[kv[0]] = kv[1]
    }
    return nil
}

func PathMap(dir string) string {
    tdir := dir
    for {
        p, ok := pathMap[tdir]
        if ok {
            p = filepath.Join(p, strings.TrimPrefix(dir, tdir))
            return p
        }
        if tdir == utils.RootPath() {
            break
        }
        tdir = filepath.Dir(tdir)
    }
    return dir
}
