package main

import (
	"virtugo/internal/dao"
	"virtugo/internal/sever"
)

func main() {
	dao.InitChromemDB()
	dao.InitSqlite()
	sever.StartSever("127.0.0.1", "8080")

}
