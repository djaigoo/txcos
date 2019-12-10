// Package xerror xerror

package xerror

import "github.com/pkg/errors"

var (
    // common
    ErrOK = errors.New("OK")
    ErrReadFile = errors.New("read file error")
    
    // conf
    ErrNotFoundRoot = errors.New("not found txcos conf dir")
    ErrReadConf     = errors.New("read conf error")
    
    // command
    ErrNoCmd       = errors.New("not recv command")
    ErrCmdNotExist = errors.New("command not exist")
    
    // cos
    ErrCosPut = errors.New("put error")
    ErrCosDel = errors.New("delete error")
)
