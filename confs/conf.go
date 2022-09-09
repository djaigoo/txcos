// Package confs

package confs

import (
    _ "embed"
    "fmt"
    "io/ioutil"
    "path/filepath"
    "sort"
    "strings"
    
    "github.com/djaigoo/txcos/xerror"
    "github.com/pkg/errors"
    "gopkg.in/yaml.v3"
)

//go:embed default_conf.yaml
var DefaultConf []byte

var yamlConf *yamlConfig

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

type ignore struct {
    Dirs  []string `yaml:"dirs"`
    Files []string `yaml:"files"`
}

type yamlConfig struct {
    Store     store     `yaml:"store"`
    Paths     []path    `yaml:"paths"`
    Clipboard clipboard `yaml:"clipboard"`
    Ignore    ignore    `yaml:"ignore"`
}

func YamlConf() *yamlConfig {
    if yamlConf != nil {
        return yamlConf
    }
    rootPath = RootPath()
    if rootPath == "" {
        panic(xerror.ErrNotFoundRoot)
    }
    yamlConf = &yamlConfig{}
    cnt, err := ioutil.ReadFile(SysConf())
    if err != nil {
        panic(errors.Wrap(err, "read system conf file"))
    }
    err = yaml.Unmarshal(cnt, yamlConf)
    if err != nil {
        panic(errors.Wrap(err, "yaml unmarshal"))
    }
    for i, tp := range yamlConf.Paths {
        // 绝对路径
        yamlConf.Paths[i].Path, err = filepath.Abs(filepath.Join(RootPath(), tp.Path))
        if err != nil {
            panic("invalid path: " + tp.Path + " " + err.Error())
        }
    }
    // 长路径在前
    sort.Slice(yamlConf.Paths, func(i, j int) bool {
        return yamlConf.Paths[i].Path > yamlConf.Paths[j].Path
    })
    return yamlConf
}

// CosPathAbs 获取完整cos路径
func CosPathAbs(p string) string {
    return fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com%s", yamlConf.Store.Bucket,
        yamlConf.Store.AppID, yamlConf.Store.Region, filepath.Join("/", p))
}

// CosPathMap 传入本地路径 返回cos地址
func CosPathMap(p string) string {
    p, err := filepath.Abs(p)
    if err != nil {
        panic("invalid path: " + p + " " + err.Error())
    }
    for _, tp := range yamlConf.Paths {
        if strings.HasPrefix(p, tp.Path) {
            return filepath.Join("/", tp.Redirect, strings.TrimPrefix(p, tp.Path))
        }
    }
    return ""
}

// CosPathClipboard 获取剪贴板cos上传路径
func CosPathClipboard(s string) string {
    if yamlConf.Clipboard.Domain != "" {
        return filepath.Join(yamlConf.Clipboard.Domain, yamlConf.Clipboard.Path, s)
    }
    return filepath.Join("/", yamlConf.Clipboard.Path, s)
}

// CosPathAbsClipboard 获取剪贴板cos上传完整url
func CosPathAbsClipboard(s string) string {
    if yamlConf.Clipboard.Domain != "" {
        return filepath.Join(yamlConf.Clipboard.Domain, yamlConf.Clipboard.Path, s)
    }
    return CosPathAbs(filepath.Join("/", yamlConf.Clipboard.Path, s))
}
