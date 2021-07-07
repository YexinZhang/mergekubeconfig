# merge kubeconfig
合并同一目录下面的kubeconfig文件 为一个单独的文件
核心思想就是扫描指定或者当前文件夹下面的kubeconfig文件， 将其解析到结构体中，在合并这些结构体


直接go build merge.go

Usage:
  -d string // 指定kubeconfig存放的文件夹, 如果不指定，将编译好的程序放入合适文件夹中
  -ctx string // 指定合并之后的current-context
  -o 合并之后输出到外界的文件， 默认路径在当前文件夹下
