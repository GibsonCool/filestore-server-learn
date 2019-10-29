package mq

import "log"

var done chan bool

// StartConsume:监听队列，获取消息
func StartConsume(qName, cName string, callback func(msg []byte) bool) {
	//1. 通过 channel.Consume 获得消息信道
	msgs, e := channel.Consume(
		qName, //队列名称
		cName, //消费者处理器名称
		true,  //自动应答通知已收到消息
		false, //非唯一的消费者，其他消费者处理器也可以去竞争这个队列里面的消息任务
		false, // rabbitMq 不支持了，只能设置false
		false, // false 表示会阻塞直到有消息过来
		nil,
	)

	if e != nil {
		log.Println(e.Error())
		return
	}

	done = make(chan bool)

	//2.循环获取队列的消息
	go func() {
		for msg := range msgs {
			//3.调用 callback 方法来处理获取的消息
			log.Printf("接受到的任务消息：%s", msg.Body)
			procssSuc := callback(msg.Body)
			log.Printf("是否成功：%t", procssSuc)
			if !procssSuc {
				// TODO：如果任务处理失败，加入错误队列，待后续处理重试
			}
		}

	}()

	// done 没有新消息过来，则会阻塞
	<-done

	// done 收到消息说明不在监听处理消息，则关闭 rabbitMQ 信道
	_ = channel.Close()
}

func StopConsumer() {
	done <- true
}
