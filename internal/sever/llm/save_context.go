package llm

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"virtugo/internal/dao"
)

func SaveContext(role string, content string) {
	InsertMessage(dao.SqliteDB, role, content)
}
func InsertMessage(db *sql.DB, role string, content string) {
	_, err := db.Exec("INSERT INTO messages (role, content) VALUES (?, ?)", role, content)
	if err != nil {
		log.Fatal(err)
	}
}

func SaveMessageToSession(db *sql.DB, sessionID int, role string, senderName string, content string, tokenCount int) (int64, error) {
	// 验证会话是否存在
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM 会话聊天 WHERE id = ?", sessionID).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, errors.New("会话不存在")
	}

	// 计算token数（简单估算）
	if tokenCount <= 0 {
		tokenCount = len(content) / 4
		if tokenCount < 1 {
			tokenCount = 1
		}
	}

	// 插入消息
	messageID, err := dao.InsertMessage(db, sessionID, role, senderName, content, tokenCount)
	if err != nil {
		return 0, err
	}

	// 更新会话的更新时间
	_, err = db.Exec(`
		UPDATE 会话聊天
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, sessionID)
	if err != nil {
		return messageID, err // 返回消息ID，但同时返回错误
	}

	return messageID, nil
}
