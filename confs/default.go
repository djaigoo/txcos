// Package conf conf

package confs

const (
    DefaultConf = `# 配置模板
store: # 存储
    type: cos # 存储类型
    secret_id: ""
    secret_key: ""
    app_id: ""
    bucket: ""
    region: ""
paths: # 上传路径
    - path: . # 本地路径
    redirect: "/txcos" # 上传路径重定向
    - path: confs # 本地路径
    redirect: "/abc"
clipboard:
    path: clipboard
`
)
