// Package conf conf

package confs

const (
    DefaultConf = `# 配置模板
# cos相关配置
secret_id=""
secret_key=""
app_id=""
bucket=""
region=""

# 路由映射
# 路由映射，格式"public:/,source:/private"，
# 前路径表示对于配置文件目录的父目录的相对路径，
# 后路径表示对于cos存储的路由路径
map_path="public:/,source:/private"

# 默认操作文件夹，默认是当前目录
default_path="."
`
)
