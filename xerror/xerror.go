// Package xerror xerror

package xerror

import (
    "fmt"
)

var (
    ErrNotFoundRoot = NewXError(1, "not found txcos conf dir")
    
    ErrCmdNotExist = NewXError(2, "command not exist")
)

type XError struct {
    parent error
    code   int
    msg    string
}

func NewXError(code int, format string, args ...interface{}) *XError {
    return &XError{
        parent: nil,
        code:   code,
        msg:    fmt.Sprintf(format, args...),
    }
}

// findRoot
func (e *XError) findRoot() *XError {
    err := e.parent
    for err != nil {
        te, ok := err.(*XError)
        if !ok {
            return e
        }
        e = te
    }
    return e
}

func Wrap(err error, format string, args ...interface{}) *XError {
    return &XError{
        parent: err,
        code:   0,
        msg:    fmt.Sprintf(format, args...),
    }
}

// Wrap
func (e *XError) Wrap(format string, args ...interface{}) *XError {
    return Wrap(e, format, args...)
}

func Equal(e1, e2 error) bool {
    if e1 == e2 {
        return true
    }
    xe1, ok1 := e1.(*XError)
    xe2, ok2 := e2.(*XError)
    if !(ok1 && ok2) {
        return false
    }
    te1 := xe1.findRoot()
    te2 := xe2.findRoot()
    return te1.code == te2.code && te1.msg == te2.msg
}

// Equal
func (e *XError) Equal(err error) bool {
    return Equal(e, err)
}

// Error
func (e *XError) Error() string {
    if e.parent == nil {
        return e.msg
    }
    return fmt.Sprintf("%s: %s", e.msg, e.parent.Error())
}
