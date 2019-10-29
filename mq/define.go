package mq

import "filestore-server/store"

// TransferData: 转移队列中消息载体的结构格式
type TransferData struct {
	FileHash      string
	CurlLocation  string          // 文件临时存储的地址
	DestLocation  string          // 要转移的目标地址
	DestStoreType store.StoreType //文件存储类型
}
