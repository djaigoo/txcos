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
// func NewCos(c *confs.Conf) *cosImp {
//     link := fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", c.Bucket, c.AppId, c.Region)
//     u, _ := url.Parse(link)
//     b := &cos.BaseURL{BucketURL: u}
//     return &cosImp{
//         cli: cos.NewClient(b, &http.Client{
//             // Timeout: 5 * time.Second,
//             Transport: &cos.AuthorizationTransport{
//                 SecretID:  c.SecretId,
//                 SecretKey: c.SecretKey,
//             },
//         }),
//     }
// }

func NewCos() *cosImp {
    link := fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", confs.YamlConf.Store.Bucket,
        confs.YamlConf.Store.AppID, confs.YamlConf.Store.Region)
    u, _ := url.Parse(link)
    b := &cos.BaseURL{BucketURL: u}
    return &cosImp{
        cli: cos.NewClient(b, &http.Client{
            // Timeout: 5 * time.Second,
            Transport: &cos.AuthorizationTransport{
                SecretID:  confs.YamlConf.Store.SecretID,
                SecretKey: confs.YamlConf.Store.SecretKey,
            },
        }),
    }
}

// Head 请求远端文件信息
func (c *cosImp) Head(ctx context.Context, name string) (header http.Header, err error) {
    rsp, err := c.cli.Object.Head(ctx, name, nil)
    if err != nil {
        return nil, err
    }
    rsp.Body.Close()
    return rsp.Header, nil
}

// GetFileMD5 获取远端文件MD5
func (c *cosImp) GetFileMD5(ctx context.Context, name string) (md5 string, err error) {
    header, err := c.Head(ctx, name)
    if err != nil {
        return "", err
    }
    etag := header.Get("Etag")
    return etag[1 : len(etag)-1], nil
}

// Get 获取远端文件内容
func (c *cosImp) Get(ctx context.Context, name string) (msg []byte, err error) {
    rsp, err := c.cli.Object.Get(ctx, name, nil)
    if err != nil {
        return nil, err
    }
    msg, _ = ioutil.ReadAll(rsp.Body)
    rsp.Body.Close()
    return msg, nil
}

// Put 上传内容值远端
func (c *cosImp) Put(ctx context.Context, name string, reader io.Reader) (err error) {
    resp, err := c.cli.Object.Put(ctx, name, reader, nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    fmt.Println(string(data))
    return err
}

// Delete 删除远端文件
func (c *cosImp) Delete(ctx context.Context, name string) (err error) {
    _, err = c.cli.Object.Delete(ctx, name)
    return err
}

var GClient *cosImp

func Init() {
    GClient = NewCos()
}
