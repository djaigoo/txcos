package utils

import (
    "crypto/md5"
    "fmt"
)

// GetMD5 获取MD5
func GetMD5(data []byte) string {
    gen := md5.New()
    gen.Write(data)
    return fmt.Sprintf("%x", gen.Sum(nil))
}
