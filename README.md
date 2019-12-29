# go-xxl-job-client
## xxl-job go客户端版
##介绍
xxj-job是一个Java实现的轻量级分布式任务调度平台，具体实现与使用请参考[https://github.com/xuxueli/xxl-job][1]，原版执行器亦要求Java平台，但公司部分项目是golang开发，所有自己实现了go版本的执行器。
##写在前面
* 源码中admin与执行器的rpc通讯使用的是xuxueli开发的xxl-rpc-core(参考[https://github.com/xuxueli/xxl-rpc][3])，其中com.xxl.rpc.remoting.invoker.reference.XxlRpcReferenceBean这个类的getObject里面有一行代码. 

```
		XxlRpcResponse xxlRpcResponse = futureResponse.get(timeout, TimeUnit.MILLISECONDS);
		if (xxlRpcResponse.getErrorMsg() != null) {
				throw new XxlRpcException(xxlRpcResponse.getErrorMsg());
		}
```
&emsp;&emsp;&emsp;&emsp;此处返回msg为字符串，不为空则抛异常调用失败，在golang中字符串默认是空字符不是&emsp;&emsp;&emsp;&emsp;空，肯定抛异常，所有需要对此项目的源码进行改动，改动后.  

```
		XxlRpcResponse xxlRpcResponse = futureResponse.get(timeout, TimeUnit.MILLISECONDS);
		if (StringUtils.hasLength(xxlRpcResponse.getErrorMsg())) {
				throw new XxlRpcException(xxlRpcResponse.getErrorMsg());
		}
```
&emsp;&emsp;&emsp;&emsp;改动源码后发布到你自己的mvn私服中，修改xxl-job-admin中所使用版本为你修改后的版本,接下来就开以开始admin与执行器的开发部署了。（强调一下此步不可省略，我们自己平时写代码时字符串操作最好也不要使用 != null来判空哦）

* 我所实现的go客户端执行器依赖gin的web支持（请参考：[https://github.com/gin-gonic/gin][4]）,如果你使用的web框架不是gin，请自行fork源码扩展.
* 整个设计实现是参考xxl-job-core的源码实现了go版本，核心在于admin与执行器的rpc通讯采用的序列化方式是hessian2，所有借用了apache实现的dubbo-go-hessian2（参考[https://github.com/apache/dubbo-go-hessian2][2]）。
##部署xxl-job-admin
切记先按上面的方法改动xxl-job-admin源码的pom文件，替换了xxl-rpc-core之后修改配置文件配置好mysql，启动admin。（详细步骤请参考[https://github.com/xuxueli/xxl-job][1]， 此处不要描述admin的部署。）

##部署xxl-job执行器（go版本）
###（1）引入go客户端依赖
```
go get github.com/feixiaobo/go-xxl-job-client
```
### (2) 在main方法中注册执行器，任务，启动web端口
####（1）实现任务
```
func XxlJobTest(ctx context.Context) error {
	logger.Info(ctx, "golang job run success >>>>>>>>>>>>>>")
	logger.Info(ctx, "the input param:", xxl.GetParam(ctx, "name"))
	return nil
}
```
###(2) 注册执行，任务，启动项目
```
	ginServer := server.GinSupport() //构造gin engine
	xxl.InitExecutor([]string{"http://127.0.0.1:8080/xxl-job-admin"}, "", "test-job-executor", 8081) //注册执行器
	xxl.RegisterJob("xxlGoTestJob", task.XxlJobTest) //注册任务
	ginServer.Run("0.0.0.0:8081") //启动web端口
```
* 注册执行器时的name是xxl-job-admin后台添加执行器时的name
* 注册任务时的名字是xxl-job-admin后台新增任务时的JobHandler

###（3）在xxl-job-admin后台管理页面添加执行器
![](/Users/feixiaobo/Desktop/1577631644200.jpg)

* appName为客户注册执行器时的名字
* 注册方式选择自动注册
### (4)在xxl-job-admin后台管理页面添加任务
![](/Users/feixiaobo/Desktop/1577631684132.jpg)
* JobHandler为注册任务时的name
* 执行器选择刚刚添加的执行器
* 运行模式默认BEAN模式

添加完成后启动在任务管理菜单中查看任务
![](/Users/feixiaobo/Desktop/1577632360005.jpg)

### （4）日志输出及参数传递
* go-xxl-job-client自己实现了日志输出，使用github.com/feixiaobo/go-xxl-job-client/logger包输出日志，因为golang不支持像Java的ThreadLocal一样的线程变量，已无法获取到golang的携程id,所以日志输出依赖的内容已存到context上下文遍历中，故log需要使用context变量。可参考任务配置中的日志输出,  
```
	logger.Info(ctx, "golang job run success >>>>>>>>>>>>>>")
```

* 任务参数传递，可使用xxl.GetParam获取到任务配置或执行时手动添加的参数
```
	logger.Info(ctx, "the input param:", xxl.GetParam(ctx, "name"))
```
在调度日志中点击执行日志查看任务执行日志，
![](/Users/feixiaobo/Desktop/1577632464114.jpg)

[1]: https://github.com/xuxueli/xxl-job	
[2]: https://github.com/apache/dubbo-go-hessian2
[3]: https://github.com/xuxueli/xxl-rpc
[4]: https://github.com/gin-gonic/gin