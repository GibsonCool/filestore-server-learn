package store

// 存储类型（表示文件存在哪里）
type StoreType int

const (
	_ StoreType = iota

	// 表示存储本地
	StoreLocal

	// 表示存储 阿里 oss
	StoreOss

	// 表示所有类型都存储一份数据
	StoreAll
)
