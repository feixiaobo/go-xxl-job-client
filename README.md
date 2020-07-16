# go-xxl-job-client

## xxl-job go 客户端版

## 介绍

xxj-job 是一个 Java 实现的轻量级分布式任务调度平台，具体实现与使用请参考[https://github.com/xuxueli/xxl-job][1]，原版执行器亦要求 Java 平台，但公司部分项目是 golang 开发，所以自己实现了 go 版本的执行器。

## 写在前面

- 我所实现的 go 客户端执行器 rpc 通信采用 dubbo-go 所用的类型 Java netty 的自研通信框架 getty（请参考：[https://github.com/dubbogo/getty][4]）.
- 整个设计实现是参考 xxl-job-core 的源码实现了 go 版本，核心在于 admin 与执行器的 rpc 通讯采用的序列化方式是 hessian2，所有借用了 apache 实现的 dubbo-go-hessian2（参考[https://github.com/apache/dubbo-go-hessian2][2]）。
- 支持了 shell, python, php, js, powershell，暂不支持动态编译的 groovy 模式。
- 脚本模式的分片参数会作为启动脚本时的最后两个参数，用户参数按顺序位于分片参数之前。

## 部署 xxl-job-admin

详细步骤请参考[https://github.com/xuxueli/xxl-job][1]， 此处不再描述 admin 的部署。

## 部署 xxl-job 执行器（go 版本）

### (1) 引入 go 客户端依赖

```
go get github.com/feixiaobo/go-xxl-job-client/v2
```

### (2) 在 main 方法中构建客户端 client，注册任务，启动端口

#### (1) 实现任务

```
func XxlJobTest(ctx context.Context) error {
	logger.Info(ctx, "golang job run success >>>>>>>>>>>>>>")
	logger.Info(ctx, "the input param:", xxl.GetParam(ctx, "name"))
	return nil
}
```

#### (2) 注册执行，任务，启动项目（可参考 example 目录）

```
	client := xxl.NewXxlClient(
	        option.WithAppName("go-example"),
		option.WithClientPort(8083),
	) //构建客户端
	client.RegisterJob("testJob", JobTest) //注册任务
	client.Run() //启动客户端
```

- 构建客户端时的 appName 是 xxl-job-admin 后台添加执行器时的 name
- 注册任务时的名字是 xxl-job-admin 后台新增任务时的 JobHandler

#### (3) 在 xxl-job-admin 后台管理页面添加执行器

![](https://github.com/feixiaobo/images/blob/master/1577631644200.jpg)

- appName 为客户注册执行器时的名字
- 注册方式选择自动注册

#### (4) 在 xxl-job-admin 后台管理页面添加任务

![](https://github.com/feixiaobo/images/blob/master/1577631684132.jpg)

- JobHandler 为注册任务时的 name
- 执行器选择刚刚添加的执行器
- 运行模式默认 BEAN 模式,可选择其他脚本模式（不支持 GLUE(Java)）

添加完成后启动在任务管理菜单中查看任务
![](https://github.com/feixiaobo/images/blob/master/1577632360005.jpg)

## 日志输出及参数传递

- go-xxl-job-client 自己实现了日志输出，使用 github.com/feixiaobo/go-xxl-job-client/v2/logger 包输出日志，因为 golang 不支持像 Java 的 ThreadLocal 一样的线程变量，已无法获取到 golang 的协程 id,所以日志输出依赖的内容已存到 context 上下文遍历中，故 log 需要使用 context 变量。可参考任务配置中的日志输出,

```
	logger.Info(ctx, "golang job run success >>>>>>>>>>>>>>")
```

- 任务参数传递，可使用 xxl.GetParam 获取到任务配置或执行时手动添加的参数，使用 xxl.GetSharding 获取到分片参数。

```
        param, _ := xxl.GetParam(ctx, "name") //获取输入参数
        logger.Info(ctx, "the input param:", param)
        shardingIdx, shardingTotal := xxl.GetSharding(ctx) //获取分片参数
        logger.Info(ctx, "the sharding param: idx:", shardingIdx, ", total:", shardingTotal)
```

在调度日志中点击执行日志查看任务执行日志。

[1]: https://github.com/xuxueli/xxl-job
[2]: https://github.com/apache/dubbo-go-hessian2
[3]: https://github.com/xuxueli/xxl-rpc
[4]: https://github.com/dubbogo/getty
