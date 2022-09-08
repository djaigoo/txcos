// Package conf conf

package confs

import (
    "fmt"
    "io/ioutil"
    "path/filepath"
    "sort"
    "strings"
    
    "github.com/BurntSushi/toml"
    "github.com/djaigoo/logkit"
    "github.com/djaigoo/txcos/xerror"
    "github.com/pkg/errors"
    "gopkg.in/yaml.v3"
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
    rootPath = RootPath()
    if rootPath == "" {
        return nil, xerror.ErrNotFoundRoot
    }
    cnt, err := ioutil.ReadFile(SysConf())
    if err != nil {
        panic(errors.Wrap(err, "read system conf file"))
    }
    ret := new(Conf)
    err = toml.Unmarshal(cnt, ret)
    if err != nil {
        return nil, errors.Wrap(err, "toml unmarshal")
    }
    if ret.SecretId == "" ||
        ret.SecretKey == "" ||
        ret.AppId == "" ||
        ret.Bucket == "" ||
        ret.Region == "" {
        return nil, errors.Wrap(nil, "secret_id or secret_key or app_id or bucket or region not allow empty string")
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
        kv[0], err = GetFileName(kv[0])
        if err != nil {
            logkit.Errorf("map_path specifies a local illegal path")
            continue
        }
        kv[1], _ = GetFileName(kv[1])
        pathMap[kv[0]] = kv[1]
    }
    return nil
}

// PathMap 获取本地路径对应的远端路径
func PathMap(dir string) string {
    tdir := dir
    for {
        p, ok := pathMap[tdir]
        if ok {
            p = filepath.Join(p, strings.TrimPrefix(dir, tdir))
            return p
        }
        if tdir == RootPath() {
            break
        }
        tdir = filepath.Dir(tdir)
    }
    return dir
}

var YamlConf *yamlConf

type store struct {
    Type      string `yaml:"type"`
    SecretID  string `yaml:"secret_id"`
    SecretKey string `yaml:"secret_key"`
    AppID     string `yaml:"app_id"`
    Bucket    string `yaml:"bucket"`
    Region    string `yaml:"region"`
}

type path struct {
    Path     string `yaml:"path"`
    Redirect string `yaml:"redirect"`
}

type clipboard struct {
    Path   string `yaml:"path"`
    Domain string `yaml:"domain"`
}

type yamlConf struct {
    Store     store     `yaml:"store"`
    Paths     []path    `yaml:"paths"`
    Clipboard clipboard `yaml:"clipboard"`
}

func CreateYamlConf() {
    InitRootPath()
}

func InitYamlConf() {
    if YamlConf != nil {
        return
    }
    rootPath = RootPath()
    if rootPath == "" {
        panic(xerror.ErrNotFoundRoot)
    }
    YamlConf = &yamlConf{}
    cnt, err := ioutil.ReadFile(SysYamlConf())
    if err != nil {
        panic(errors.Wrap(err, "read system conf file"))
    }
    err = yaml.Unmarshal(cnt, YamlConf)
    if err != nil {
        panic(errors.Wrap(err, "yaml unmarshal"))
    }
    for i, tp := range YamlConf.Paths {
        YamlConf.Paths[i].Path, err = filepath.Abs(filepath.Join(RootPath(), tp.Path))
        if err != nil {
            panic("invalid path: " + tp.Path + " " + err.Error())
        }
    }
    // 长路径在前
    sort.Slice(YamlConf.Paths, func(i, j int) bool {
        return YamlConf.Paths[i].Path > YamlConf.Paths[j].Path
    })
}

// CosPathAbs 获取完整cos路径
func CosPathAbs(p string) string {
    return fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com%s", YamlConf.Store.Bucket,
        YamlConf.Store.AppID, YamlConf.Store.Region, filepath.Join("/", p))
}

// CosPathMap 传入本地路径 返回cos地址
func CosPathMap(p string) string {
    p, err := filepath.Abs(p)
    if err != nil {
        panic("invalid path: " + p + " " + err.Error())
    }
    for _, tp := range YamlConf.Paths {
        if strings.HasPrefix(p, tp.Path) {
            return filepath.Join("/", tp.Redirect, strings.TrimPrefix(p, tp.Path))
        }
    }
    return ""
}

// CosPathClipboard 获取剪贴板cos上传路径
func CosPathClipboard(s string) string {
    if YamlConf.Clipboard.Domain != "" {
        return filepath.Join(YamlConf.Clipboard.Domain, YamlConf.Clipboard.Path, s)
    }
    return filepath.Join("/", YamlConf.Clipboard.Path, s)
}

// CosPathAbsClipboard 获取剪贴板cos上传完整url
func CosPathAbsClipboard(s string) string {
    if YamlConf.Clipboard.Domain != "" {
        return filepath.Join(YamlConf.Clipboard.Domain, YamlConf.Clipboard.Path, s)
    }
    return CosPathAbs(filepath.Join("/", YamlConf.Clipboard.Path, s))
}
