# txcos
txcos 是一个快速上传本地文件至腾讯云对象存储中，它会扫描指定目录下上次更新时间后有所更改的文件，并上传到对象存储指定的存储库中。

目前支持的命令：
```text
txcos usage:
  -h, help    show usage
  -i, init    initialize related configuration items
  -p, push    push add information
  -s, status  get status
```

* -h，获取帮助文档
* -i，初始化txcos目录
* -p，将有所修改的文件上传至腾讯云对象存储
* -s，查看当前有所更改的文件列表