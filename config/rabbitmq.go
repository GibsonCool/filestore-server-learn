package config

const (
	// AsyncTransferEnable : 是否开启文件异步转移（默认同步）
	AsyncTransferEnable = true

	// RabbitURL:rabbitmq 服务的入口url.   amqp 协议方式   guest:guest 账号密码
	RabbitURL = "amqp://guest:guest@127.0.0.1:5672/"

	// TransExchangeName: 在rabbitmq UI管理界面添加的交换器名称。 用来接受生产者发送的消息并将这些消息路由给服务器中的队列
	TransExchangeName = "uploadserver.trans"

	// TransOSSQueueName: 用于 oss 任务转移队列名. 通常取名是交换器名+队列任务名
	TransOSSQueueName = "uploadserver.trans.oss"

	// TransOSSErrQueueName:oss 转移失败后写入拧一个队列的队列名
	TransOSSErrQueueName = "uploadserver.trans.oss.err"

	// TransOSSRoutingKey: routingkey
	TransOSSRoutingKey = "oss"
)
