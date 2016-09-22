#DelayQ   
延迟消费的消息队列   


#设计目的
当前的MQ模型，基本都是基于实时性的生产、消费，但是对于一些实际的业务场景，你可能希望消息在某一个时刻被消费，或者在一个固定的时间被消费，如果你也是出于这种需求，那就没错了，就是这个。

#Quick Start
这个工具依赖于Redis，相信大家肯定都有使用过，其中Redis的安装部分省略。

代码使用Go语言编写，所以如需编译，请安装Golang的运行环境

将代码clone 到本地
```
git clone https://github.com/coraldane/DelayQ.git
```
然后切换到你的工作目录，如下
```
cd ~/go/src/github.com/coraldane/DelayQ
```
先执行编译
```
./control build  或者 go build
```
当前目录下面会新增一个名字为DelayQ 的二进制文件，如果是windows机器，文件名为DelayQ.exe   

程序运行默认读取当前目录下面的cfg.json 文件，如果没有会从cfg.example复制而来   
如果需要自己指定配置文件，请使用如下命令启动
```
./DelayQ -c /path/to/config/cfg.json >> /path/to/logfile/app.log &
```

配置文件如下:   
```
{
  "log_level": "debug",     ##日志文件级别
  "protocol": "http",       ##这里是指定提供服务的接口协议，支持http和tcp两种
  "listen": "127.0.0.1:23456",   ##绑定机器的端口
  "bucket_key_size": 10,         ##消息池的桶大小，如果你部署的DelayQ实例数很多，可以适当增大一些提高并发数
  "redis": {                ##这里是Redis的配置
      "dsn": "127.0.0.1:6379",
      "passwd": "******",
      "maxIdle": 5,
      "connTimeout": 5000,
      "readTimeout": 5000,
      "writeTimeout": 5000
  }
}
```

以上都配置好后，建议使用如下方式启动   

注：DelayQ是支持多实例部署的，直接启动多个这样的实例即可   

```
./control start
```
启动以后检查一下，机器上面的端口是否正常  

#API Reference
由于接口通过Hprose(http://www.hprose.com/)来对外提供服务，你可以开发自己语言的Client调用端   
下面以Java为例:
```
public interface DelayQueue {

	/**
	 * 发布延迟消息任务，生产者调用
	 * @param job
	 * @return 如果出错则返回错误信息，否则返回空串
	 */
	String publish(Job job);
	
	/**
	 * 消费消息, 消费者调用
	 * @param origin 源，一般为系统的唯一标识
	 * @param topic 消息的Topic
	 * @param nextSeconds  TTR超时时间，在之后的*秒后未收到ACK则会重新消费
	 * @return
	 */
	Job consume(int origin, int topic, int nextSeconds);
	
	/**
	 * 消息消费成功后，告知DelayQ，否则会重复消费
	 * @param jobId 格式为origin_topic_bizId, 中间以英文状态的下划线拼接
	 * @return 如果出错则返回错误信息，否则返回空串
	 */
	String ack(String jobId);
}
```
其中的Job定义如下:   
```
public class Job {
	private int origin;   //源，一般为系统的唯一标识
	private int topic;    //消息的Topic
	private long bizId;   //消息的业务ID，同一个topic下的bizId不允许重复
	
	@JSONField(format="yyyy-MM-dd HH:mm:ss")
	private Timestamp gmtNext; //希望消息被执行的下一时间
	private String body;       //消息体内容
	private int retryCount;    //消息被消费次数
}
```
