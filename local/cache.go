package local

var SFileList []File

// Cache
type Cache struct {
    Create []File `json:"create"`
    Delete []File `json:"delete"`
    Modify []File `json:"modify"`
}

// NewRecord new Cache
func NewRecord(sf *ScanFile) *Cache {
    return &Cache{}
}
