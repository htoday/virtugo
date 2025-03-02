package dao

import (
	"github.com/philippgille/chromem-go"
	"log"
)

var ChromemDB *chromem.DB

func InitChromemDB() {
	var err error
	ChromemDB, err = chromem.NewPersistentDB("./chromem.db", false)
	if err != nil {
		log.Println("创建chromem向量数据库失败", err)
	}

}
