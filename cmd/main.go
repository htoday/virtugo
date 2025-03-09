package main

import (
	"virtugo/internal/config"
	"virtugo/internal/dao"
	"virtugo/internal/sever"
)

func main() {
	dao.InitChromemDB()
	dao.InitSqlite()
	config.LoadConfig()
	sever.StartSever("127.0.0.1", "8080")
}
