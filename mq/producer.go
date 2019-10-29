package mq

import (
	"filestore-server/config"
	mySql "filestore-server/db/mysql"
	"github.com/streadway/amqp"
	"log"
)

var conn *amqp.Connection
var channel *amqp.Channel

// 用于 rabbitMq 异常关闭接受通知的 chan
var notifyClose chan *amqp.Error

func init() {
	// 只有开启异步转移功能才初始化 rabbitMq 链接
	if !config.AsyncTransferEnable {
		return
	}

	if initChannel() {
		channel.NotifyClose(notifyClose)
	}

	mySql.Setup()

	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				log.Printf("rabbitMQ 异常退出了, msg:%v，尝试重连中...\n", msg)
				initChannel()
			}
		}
	}()
}

// 初始化获取连接
func initChannel() bool {
	// 1.判断 channel 是否已经创建过
	if channel != nil {
		return true
	}

	// 2.获取 rabbitmq 的一个链接
	var e error
	conn, e = amqp.Dial(config.RabbitURL)
	if e != nil {
		log.Println(e.Error())
		return false
	}

	// 3.打开一个 channel 用于消息的发布与接收
	channel, e = conn.Channel()
	if e != nil {
		log.Println(e.Error())
		return false
	}
	return true
}

func Publish(exchange, routingkey string, msg []byte) bool {
	// 1.判断 channel 是否正常
	if !initChannel() {
		return false
	}
	// 2.执行消息发布动作
	e := channel.Publish(
		exchange,
		routingkey,
		false, // 如果没有合适的的队列转发的话，交换机是否将消息返回给发布者 ,false 不返回则消息丢失
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		},
	)

	if e != nil {
		log.Println(e.Error())
		return false
	}

	return true
}
