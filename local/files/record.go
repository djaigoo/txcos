// Package Files Files

package files

import (
    "encoding/json"
    "io/ioutil"
    "path/filepath"
    "reflect"
    "sort"
    "time"
    "unsafe"
    
    "github.com/djaigoo/logkit"
    "github.com/djaigoo/txcos/utils"
)

// File
type File struct {
    Name       string    `json:"name"`
    Dir        string    `json:"dir"`
    ModifyTime time.Time `json:"modify_time"` // system stat time
    UpdateTime time.Time `json:"update_time"` // cos update time
}

func (f File) FilePath() string {
    return filepath.Join(f.Dir, f.Name)
}

func NewFile(filename string, mtime time.Time) File {
    name := filepath.Base(filename)
    dir := filepath.Dir(filename)
    return File{
        Name:       name,
        Dir:        dir,
        ModifyTime: mtime,
    }
}

func Compare(a, b File) int {
    if a.Dir < b.Dir {
        return -1
    }
    if a.Dir > b.Dir {
        return 1
    }
    if a.Name < b.Name {
        return -1
    }
    if a.Name > b.Name {
        return 1
    }
    return 0
}

func Sort(filelist []File) {
    sort.Slice(filelist, func(i, j int) bool {
        if Compare(filelist[i], filelist[j]) < 0 {
            return true
        }
        return false
    })
}

func Index(fs []File, f File) (idx int) {
    for i := range fs {
        if Compare(fs[i], f) == 0 {
            return i
        }
    }
    return -1
}

func Merge(fs *[]File, crt, mod, del []File) {
    if len(del) != 0 {
        Sort(del)
        ifs, idel := 0, 0
        for ifs+idel < len(*fs) {
            if idel < len(del) && Compare((*fs)[ifs+idel], del[idel]) == 0 {
                idel++
                continue
            }
            (*fs)[ifs] = (*fs)[ifs+idel]
            ifs++
        }
        (*reflect.SliceHeader)(unsafe.Pointer(fs)).Len = ifs
    }
    tn := time.Now()
    for _, f := range mod {
        idx := Index(*fs, f)
        if idx == -1 {
            logkit.Errorf("index error %s", f.FilePath())
            continue
        }
        (*fs)[idx].UpdateTime = tn
    }
    
    for _, f := range crt {
        f.UpdateTime = tn
        *fs = append(*fs, f)
    }
    Sort(*fs)
}

var GFileList []File

func InitRecord() {
    GFileList = make([]File, 0)
    data, err := ioutil.ReadFile(utils.SysRecord())
    if err != nil {
        panic("read record file error " + err.Error())
        return
    }
    err = json.Unmarshal(data, &GFileList)
    if err != nil {
        panic("json unmarshal error " + err.Error())
        return
    }
}

func CloseRecord() (err error) {
    data, err := json.Marshal(GFileList)
    if err != nil {
        return err
    }
    return ioutil.WriteFile(utils.SysRecord(), data, 0644)
}
