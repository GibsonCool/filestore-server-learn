module filestore-server

go 1.13

require (
	github.com/aliyun/aliyun-oss-go-sdk v0.0.0-20190307165228-86c17b95fcd5
	github.com/garyburd/redigo v1.6.0
	github.com/gin-gonic/gin v1.4.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.3.2
	github.com/json-iterator/go v1.1.7
	github.com/micro/go-micro v1.14.0
	github.com/micro/go-plugins v1.4.0 // indirect
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
)

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.0
