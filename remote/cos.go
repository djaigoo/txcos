// Package cos cos

package remote

import (
    "context"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    
    "github.com/djaigoo/txcos/confs"
    "github.com/tencentyun/cos-go-sdk-v5"
)

// cosImp
type cosImp struct {
    cli *cos.Client
}

// NewCos new cosImp
func NewCos(c *confs.Conf) *cosImp {
    link := fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", c.Bucket, c.AppId, c.Region)
    u, _ := url.Parse(link)
    b := &cos.BaseURL{BucketURL: u}
    return &cosImp{
        cli: cos.NewClient(b, &http.Client{
            // Timeout: 5 * time.Second,
            Transport: &cos.AuthorizationTransport{
                SecretID:  c.SecretId,
                SecretKey: c.SecretKey,
            },
        }),
    }
}

func (c *cosImp) Head(ctx context.Context, name string) (header http.Header, err error) {
    rsp, err := c.cli.Object.Get(ctx, name, nil)
    if err != nil {
        return nil, err
    }
    rsp.Body.Close()
    return rsp.Header, nil
}

func (c *cosImp) GetFileMD5(ctx context.Context, name string) (md5 string, err error) {
    header, err := c.Head(ctx, name)
    if err != nil {
        return "", err
    }
    etag := header.Get("Etag")
    return etag[1 : len(etag)-1], nil
}

// Get Get
func (c *cosImp) Get(ctx context.Context, name string) (msg []byte, err error) {
    rsp, err := c.cli.Object.Get(ctx, name, nil)
    if err != nil {
        return nil, err
    }
    msg, _ = ioutil.ReadAll(rsp.Body)
    rsp.Body.Close()
    return msg, nil
}

// Put Put
func (c *cosImp) Put(ctx context.Context, name string, reader io.Reader) (err error) {
    _, err = c.cli.Object.Put(ctx, name, reader, nil)
    return err
}

// Delete Delete
func (c *cosImp) Delete(ctx context.Context, name string) (err error) {
    _, err = c.cli.Object.Delete(ctx, name)
    return err
}

var GClient *cosImp

func Init() {
    GClient = NewCos(confs.GCos)
}
