## Xunray Scaffold
---
对 Gin 的进一步封装
1. 服务注册
2. 服务客户端
3. 封装了 Storage 存储
4. 协议自绑定
5. 接口单元测试

## 依赖
> 需要安装 Consul，保证能访问：[http://127.0.0.1:8500](http://127.0.0.1:8500)

## 使用方式
---
1. 目录结构
![UTOOLS1571880728610.png](https://i.loli.net/2019/10/24/xQu1DEij5MSbaBH.png)
- main.go 入口文件
- server
    - packages 核心逻辑目录
    - handlers 控制器
    - consts 常量定义和枚举
    - models 数据库模型
    - protocols 协议
2. 配置信息
> 配置信息是由 viper 管理的，配置文件格式：yaml
```yaml
runmode: debug   # 开发模式, debug, release, test

server:
#  addr: 0.0.0.0
#  port: 5001

db:
  bmp:
    master:
      host: "127.0.0.1:3307"
      name: "xunray_bmp"
      user: "root"
      pswd: "111111"
    openConns: 20
    idleConns: 20
    showLog: true
```
**注意：server.addr、server.port 如果不指定，则会获取当前所在机器的内网地址，且随机选择一个可用的端口**

> 指定配置文件信息到环境变量，以 Linux 为例：
```bash
export BMP_CONFIG=/Users/xiangzhi/Work/Go/src/code.xunray.com/bmp/conf/config.yaml
```
3. 基本用法
```go
package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/xinzf/xunray"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"code.xunray.com/bmp/server/handlers"
	"code.xunray.com/bmp/server/middleware"
	"code.xunray.com/bmp/server/protocols"
)

func init() {
	// 初始化 Servere
	// 注意：入参是该服务的名称，和上面环境变量的名字要对应（BMP_CONFIG）
	xunray.Server.Init("BMP")

    // 装载路由插件
	xunray.Server.Use(middleware.Logger())

    // 自定义错误响应
	xunray.Server.ErrorHandler(func(err error) interface{} {
		return protocols.Response{
			Code:    -1000,
			Message: err.Error(),
		}
	})

    // 注册服务
	xunray.Server.Register("bmp.app.list", new(handlers.App).List)
	xunray.Server.Register("bmp.app.create", new(handlers.App).Create)
	xunray.Server.Register("bmp.app.update", new(handlers.App).Update)
	xunray.Server.Register("bmp.app.delete", new(handlers.App).Delete)
	xunray.Server.Register("bmp.app.copy", new(handlers.App).Copy)

	xunray.Server.Register("bmp.form.list", new(handlers.Form).List)
	xunray.Server.Register("bmp.form.create", new(handlers.Form).Create)
	xunray.Server.Register("bmp.form.detail", new(handlers.Form).Detail)
	xunray.Server.Register("bmp.form.update", new(handlers.Form).Update)
	xunray.Server.Register("bmp.form.layout.save", new(handlers.Form).SaveLayout)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

    // 启动服务
	go xunray.Server.Start()

	<-ch

	if err := xunray.Server.Stop(); err != nil {
		log.Errorln(err)
	}
}
```

## Storage 用法
> Mysql
```go
storage.DB.Use(DBNAME)
```

> Redis
```go
storage.Redis.Client()
```

> Mongo
```go
storage.Mongo.Use(DBNAME)
```

## Httpclient 用法
```go
// 先定义一个 struct 实现 httpclient.Requester 接口
type TestRequest struct {
}

// 发起请求前要做的准备事情，比如一些前期验证可以在这里做，如果返回 error ，则阻止请求的发送
func (this *TestRequest) Prepare() error{
	
}

// 获取请求地址
func (this *TestRequest) GetURI() string{
	
}

// 获取 POST或者PUT的数据，只有在 method=post|put的时候才有效
func (this *TestRequest) GetPostData() []byte{
	
}

// 获取要发送的头信息
func (this *TestRequest) GetHeaders() map[string]string{
	
}

// 获取请求方式：GET、POST、PUT、DELETE
func (this *TestRequest) GetMethod() string{
	
}

// 响应回调
func (this *TestRequest) Handle(rsp []byte, httpStatus int, err error){
	
}

// 获取请求错误
func (this *TestRequest) Error() error{
	
}

// new 一个 httpclient 对象出来，然后把这个 request 塞给这个 client
client:=httpclient.New()
	
request:=new(TestRequest)
// 设置该 request
// ...

client.AddRequest(request)
if err := client.Exec();err!=nil{
    // ...
}

if request.Error() != nil {
    // ...
}

// ...
```

## ServiceClient 用法
> ServiceClient 是用于请求其他服务的客户端
```go
var rsp map[string]interface{}
if err := xunray.Client.Call("service name",[]byte("body"),&rsp);err!=nil{
    // ...	
}	
fmt.Printf("%+v",rsp)
```