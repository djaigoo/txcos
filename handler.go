package main

import (
    "sort"
    "strings"
    "unicode"
    
    "github.com/djaigoo/txcos/xerror"
    "github.com/pkg/errors"
)

type handler func() error

var handlerMap map[string]handler
var usageMap map[string]string

func init() {
    handlerMap = make(map[string]handler, 10)
    usageMap = make(map[string]string, 10)
}

func register(cmd string, usage string, handle handler) {
    usageMap[cmd] = usage
    rcmd := ""
    for _, c := range cmd {
        if !unicode.IsSpace(c) {
            rcmd += string(c)
        }
    }
    cmds := strings.Split(rcmd, ",")
    for _, c := range cmds {
        handlerMap[c] = handle
    }
}

func do(cmd string) error {
    handler, ok := handlerMap[cmd]
    if !ok {
        return errors.Wrapf(xerror.ErrCmdNotExist, "command %s", cmd)
    }
    return handler()
}

func usage() string {
    maxlen := 0
    keys := make([]string, 0, len(usageMap))
    for k := range usageMap {
        if len(k) > maxlen {
            maxlen = len(k)
        }
        keys = append(keys, k)
    }
    
    sort.Slice(keys, func(i, j int) bool {
        return keys[i] < keys[j]
    })
    
    str := strings.Builder{}
    str.WriteString("txcos usage:\n")
    for _, k := range keys {
        str.WriteString("  ")
        diff := maxlen - len(k) + 2
        str.WriteString(k)
        for i := 0; i < diff; i++ {
            str.WriteByte(' ')
        }
        str.WriteString(usageMap[k])
        str.WriteByte('\n')
    }
    
    return str.String()
}
