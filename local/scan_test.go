// Package Files Files

package local

import (
    "fmt"
    "testing"
    "time"
)

func newScanFile() *ScanFile {
    InitIgnore()
    sf := NewScanFile()
    return sf
}

func TestWalk(t *testing.T) {
    a := map[int]int{1: 2, 3: 4}
    b := 2
    b = a[4]
    fmt.Println(b)
    fmt.Println(time.Now().Unix())
    return
    
    go func() {
        if err := recover(); err != nil {
            fmt.Println(err)
        }
    }()
    sf := newScanFile()
    // go func() {
    //     for range time.NewTicker(100 * time.Millisecond).C {
    //         fmt.Println(len(Files.Files))
    //     }
    // }()
    start := time.Now()
    path := "/Users/daihao/github"
    err := sf.Walk(path)
    if err != nil {
        return
    }
    fmt.Printf("%s\n", time.Now().Sub(start))
    fmt.Println(len(sf.Files))
    // fs := record.Files{Files: sf.Files}
    // data, _ := json.Marshal(fs)
    
    // fmt.Println(string(data))
    // fmt.Printf("%#v", sf.Files)
}
