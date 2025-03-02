package dao

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var SqliteDB *sql.DB

func InitSqlite() {
	// 连接 SQLite 数据库（如果不存在会自动创建）
	var err error
	SqliteDB, err = sql.Open("sqlite3", "memory.db")
	if err != nil {
		log.Fatal("连接数据库失败", err)
	}
	//defer SqliteDB.Close()
	_, err = SqliteDB.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		role TEXT,
		content TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal("数据库建表失败", err)
	}

}
