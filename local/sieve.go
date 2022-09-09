package local

import (
    "path/filepath"
    
    "github.com/djaigoo/txcos/confs"
    "github.com/djaigoo/txcos/xerror"
)

// sieveImp 默认筛选路径，只操作里面包含的路径
type sieveImp struct {
    paths []string
}

func (si *sieveImp) Add(paths ...string) (err error) {
    for _, path := range paths {
        if !filepath.IsAbs(path) {
            path, err = filepath.Abs(path)
            if err != nil {
                return xerror.Wrap(err, "abs path")
            }
        }
        si.paths = append(si.paths, path)
    }
    return nil
}

// Find
func (si *sieveImp) Find(path string) bool {
    for _, p := range si.paths {
        if filepath.HasPrefix(path, p) {
            return true
        }
    }
    return false
}

// Paths
func (si *sieveImp) Paths() []string {
    return si.paths
}

var sieve *sieveImp

func Sieve() *sieveImp {
    if sieve == nil {
        sieve = &sieveImp{}
        for _, p := range confs.YamlConf().Paths {
            sieve.Add(p.Path)
        }
    }
    return sieve
}
