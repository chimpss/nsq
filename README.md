### before learn

- nsq的项目使用go-mod搭建
- 出了apps，其他的所有目录看成一个包
- 所有的程序入口放到apps下
- fmt.sh：goimports当前目录下所有文件
- coverage.sh: 测试相关文件，测试覆盖率
- dist.sh: 构建发布文件

### apps

- 所有的可执行程序的入口。
- 守护进程的入口

### bench

- 这个还没有细看，看字面意思应该是并发test

### contrib

- 配置文件

### internal

- 各种用到的工具包放在了这
- 比如封装了tcp相关函数，topic正则匹配的`protocal`包
- 比如自定义的log包`lg`，虽然小，但是支持了`log level`，可以根据配置选择性打印log
- 比如封装了http相关的函数，封装了`gzip`的`http_api`包，这个看了一部分
- 比如权限认证的包`auth`，统一了整个nsq的权限规则，这个还没有细看
- 剩下的一些工具库还没有看完

### nsqadmin

- 就是可视化工具，实时监控

### nsqd

-  是一个守护进程，负责接收，排队，投递消息给客户端。
- ![topics/channels](http://wiki.jikexueyuan.com/project/nsq-guide/images/internal1.gif)

### nsqlookupd

- 是守护进程负责管理拓扑信息。客户端通过查询 nsqlookupd 来发现指定话题（topic）的生产者，并且 nsqd 节点广播话题（topic）和通道（channel）信息。
- 有两个接口：TCP 接口，nsqd 用它来广播。HTTP 接口，客户端用它来发现和管理。



### 拓扑模式





