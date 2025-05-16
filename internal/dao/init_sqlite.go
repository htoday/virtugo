package dao

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"virtugo/logs"
)

var SqliteDB *sql.DB

func InitSqlite() {
	// 连接 SQLite 数据库（如果不存在会自动创建）
	var err error
	SqliteDB, err = sql.Open("sqlite3", "memory.db")
	if err != nil {
		log.Fatal("连接数据库失败", err)
	}
	schema := `
	CREATE TABLE IF NOT EXISTS 用户 (
		username CHAR VARYING PRIMARY KEY,
		password CHAR VARYING NOT NULL,
		qq_id CHAR VARYING
	);
	CREATE TABLE IF NOT EXISTS 会话聊天 (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username CHAR VARYING NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		title CHAR VARYING NOT NULL,
		FOREIGN KEY (username) REFERENCES 用户(username)
	);
	CREATE TABLE IF NOT EXISTS 消息 (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conversation_id INTEGER NOT NULL,
		role CHAR VARYING NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		token_count INTEGER NOT NULL,
		content CHAR VARYING NOT NULL,
		senderName CHAR VARYING,
		FOREIGN KEY (conversation_id) REFERENCES 会话聊天(id)
	);
	`
	_, err = SqliteDB.Exec(schema)
	if err != nil {
		logs.Logger.Fatal("数据库建表失败" + err.Error())
	}
	//defer SqliteDB.Close()
	//_, err = SqliteDB.Exec(`CREATE TABLE IF NOT EXISTS messages (
	//	id INTEGER PRIMARY KEY AUTOINCREMENT,
	//	role TEXT,
	//	content TEXT,
	//	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	//)`)
	//if err != nil {
	//	log.Fatal("数据库建表失败", err)
	//}

}
