package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

var db *sql.DB

func init() {
	db, _ = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/fileserver?charset=utf8")

	db.SetMaxOpenConns(1000)
	e := db.Ping()
	if e != nil {
		fmt.Printf("Failed to connect to mysql ,err :%s", e.Error())
		os.Exit(1)
	}
}

// DBConn: 返回数据库连接对象
func DBConn() *sql.DB {
	return db
}

func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)

	for rows.Next() {
		// 将行数据保存到 record 字典中
		e := rows.Scan(scanArgs...)
		checkErr(e)

		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		records = append(records, record)
	}

	return records
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
