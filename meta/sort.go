package meta

import "time"

const baseFormat = "2006-01-02 15:04:05"

/*
	实现 FileMeta 在map中自定义排序条件
	根据时间大小来排序
*/
type ByUploadTime []FileMeta

func (a ByUploadTime) Len() int {
	return len(a)
}

func (a ByUploadTime) Less(i, j int) bool {
	iTime, _ := time.Parse(baseFormat, a[i].UploadAt)
	jTime, _ := time.Parse(baseFormat, a[j].UploadAt)
	return iTime.UnixNano() > jTime.UnixNano()
}

func (a ByUploadTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
