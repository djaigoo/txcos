# txcos
txcos 是一个快速上传本地文件至腾讯云对象存储中，它会扫描指定目录下上次更新时间后有所更改的文件，并上传到对象存储指定的存储库中。

## 参数说明
目前支持的命令：
```text
txcos usage:
  -h, help    show usage
  -i, init    initialize related configuration items
  -l, pull    pull remote file
  -p, push    push add information
  -s, status  get status
```

* -h，获取帮助文档
* -i，初始化txcos目录
* -l，拉取远端文件
* -p，将有所修改的文件上传至腾讯云对象存储
* -s，查看当前有所更改的文件列表

## 配置目录说明
在上传目录中使用init新建`.cos`目录作为txcos的相关配置目录
`conf.toml`是配置文件，配置cos相关内容和txcos系统参数
支持忽略文件和文件夹，在`.cos`目录`.ignore`文件配置

## 使用方法
在需要上传的目录调用
```bash
txcos init
```
会在当前目录中生成.cos文件夹。
在conf.toml填写相关配置，可以声明默认目录，在使用时不需要手动指定目录，还可以设定本地与远端的文件夹名映射关系。
在.ignore中配置不需要上传的文件或文件夹名（文件夹名后面一点要加上`/`，不然会被认为是文件）

查看本地有哪些文件做了修改：
```bash
txcos status
```
会打印当前目录结构与上次上传时的目录结构的差距，会输出新建，修改，删除文件列表。

上传本地文件至远端：
```bash
txcos push
```
会按照当前status的结果进行同步远端文件数据。

拉取远端文件至本地：
```bash
txcos pull
```
会按照当前status的结果将本地文件恢复至远端文件内容。

status，push，pull后面都可以接上路径进行部分目录的操作。