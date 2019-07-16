package errorUtils

import "fmt"

const FailedStartServer = "Failed to start server ,err: %s"
const FailedGetData = "Failed to get data, err:%s"

const FailedCreateFile = "Failed to create file, err:%s"
const FailedSaveData = "Failed to save data into file , err:%s"

func SimplePrint(err error, formt string) {
	if err != nil {
		fmt.Printf(formt, err.Error())
	}
}
