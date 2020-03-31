# go-xxl-job-client
## xxl-job go客户端版
## 介绍
xxj-job是一个Java实现的轻量级分布式任务调度平台，具体实现与使用请参考[https://github.com/xuxueli/xxl-job][1]，原版执行器亦要求Java平台，但公司部分项目是golang开发，所有自己实现了go版本的执行器。
## 写在前面
* 我所实现的go客户端执行器rpc通信采用dubbo-go所用的类型Java netty的自研通信框架getty（请参考：[https://github.com/dubbogo/getty][4]）.
* 整个设计实现是参考xxl-job-core的源码实现了go版本，核心在于admin与执行器的rpc通讯采用的序列化方式是hessian2，所有借用了apache实现的dubbo-go-hessian2（参考[https://github.com/apache/dubbo-go-hessian2][2]）。
* 支持了shell, python, php, js, powershell，暂不支持动态编辑的groovy模式。
* 脚本模式的分片参数会作为启动脚本时的最后两个参数，用户参数按顺序位于分片参数之前。

## 部署xxl-job-admin
详细步骤请参考[https://github.com/xuxueli/xxl-job][1]， 此处不再描述admin的部署。）
## 部署xxl-job执行器（go版本）
### (1) 引入go客户端依赖
```
go get github.com/feixiaobo/go-xxl-job-client
```
### (2) 在main方法中构建客户端client，注册任务，启动端口
#### (1) 实现任务
```
func XxlJobTest(ctx context.Context) error {
	logger.Info(ctx, "golang job run success >>>>>>>>>>>>>>")
	logger.Info(ctx, "the input param:", xxl.GetParam(ctx, "name"))
	return nil
}
```
#### (2) 注册执行，任务，启动项目（可参考example目录）
```
	client := xxl.NewXxlClient(
	        option.WithAppName("go-example"),
		option.WithClientPort(8083),
	) //构建客户端
	client.RegisterJob("testJob", JobTest) //注册任务
	client.Run() //启动客户端
```
* 构建客户端时的appName是xxl-job-admin后台添加执行器时的name
* 注册任务时的名字是xxl-job-admin后台新增任务时的JobHandler
#### (3) 在xxl-job-admin后台管理页面添加执行器
![](https://github.com/feixiaobo/images/blob/master/1577631644200.jpg)
* appName为客户注册执行器时的名字
* 注册方式选择自动注册
#### (4) 在xxl-job-admin后台管理页面添加任务
![](https://github.com/feixiaobo/images/blob/master/1577631684132.jpg)
* JobHandler为注册任务时的name
* 执行器选择刚刚添加的执行器
* 运行模式默认BEAN模式,可选择其他脚本模式（不支持GLUE(Java)）

添加完成后启动在任务管理菜单中查看任务
![](https://github.com/feixiaobo/images/blob/master/1577632360005.jpg)
## 日志输出及参数传递
* go-xxl-job-client自己实现了日志输出，使用github.com/feixiaobo/go-xxl-job-client/logger包输出日志，因为golang不支持像Java的ThreadLocal一样的线程变量，已无法获取到golang的协程id,所以日志输出依赖的内容已存到context上下文遍历中，故log需要使用context变量。可参考任务配置中的日志输出,  
```
	logger.Info(ctx, "golang job run success >>>>>>>>>>>>>>")
```

* 任务参数传递，可使用xxl.GetParam获取到任务配置或执行时手动添加的参数
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
