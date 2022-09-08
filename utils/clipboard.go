package utils

import (
    "fmt"
    "time"
    
    "github.com/pkg/errors"
    "golang.design/x/clipboard"
)

// GetClipboardImage 从剪贴板中获取png图片
func GetClipboardImage() (name string, data []byte, err error) {
    err = clipboard.Init()
    if err != nil {
        return
    }
    data = clipboard.Read(clipboard.FmtImage)
    if len(data) == 0 {
        err = errors.New("not found image")
        return
    }
    name = fmt.Sprintf("%s%s.png", time.Now().Format("20060102"), GetMD5(data))
    return
}

func SetClipboardText(text []byte) {
    clipboard.Init()
    clipboard.Write(clipboard.FmtText, text)
}
