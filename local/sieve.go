package local

import (
    "path/filepath"
    
    "github.com/pkg/errors"
)

var GSieve []string

func InitSieve(paths ...string) {
    GSieve = make([]string, 0, 4)
    err := AddSieve(paths...)
    if err != nil {
        panic(err.Error())
    }
}

func AddSieve(paths ...string) (err error) {
    for _, path := range paths {
        if !filepath.IsAbs(path) {
            path, err = filepath.Abs(path)
            if err != nil {
                return errors.Wrap(err, "abs path")
            }
        }
        GSieve = append(GSieve, path)
    }
    return nil
}

func FindSieve(path string) bool {
    for i := range GSieve {
        if filepath.HasPrefix(path, GSieve[i]) {
            return true
        }
    }
    return false
}
