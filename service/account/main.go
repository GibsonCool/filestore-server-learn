package main

import (
	"filestore-server/config"
	"filestore-server/service/account/handler"
	"filestore-server/service/account/proto"
	"github.com/micro/go-micro"
	"github.com/micro/go-plugins/registry/consul"
	"log"
	"time"
)

func main() {
	config.SetupInit()
	// 创建一个 service
	service := micro.NewService(
		micro.Name(config.MicroServiceUserName),
		/*
			默认情况下,consul 不会自动移除无效的实例，所以一般需要手动设置一个超时时间（健康检查TTL)

			这个超时时间用于通知 consul,如果这段时间内相关服务没有主动报心跳信息，那么 consul 就可以
			认为该服务不正常，可以从注册表中移除。

			另外指定服务报心跳的时间间隔，一般要比超时时间短。
		*/
		micro.RegisterTTL(time.Second*10),     //10s 检查等待时间
		micro.RegisterInterval(time.Second*5), //服务每5s发一次心跳

		micro.Registry(consul.NewRegistry()),
	)

	service.Init()
	proto.RegisterUserServiceHandler(service.Server(), new(handler.User))
	if err := service.Run(); err != nil {
		log.Println("注册服务启动失败：" + err.Error())
	}
}
