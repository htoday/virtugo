package dao

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

// CreateUser creates a new user
func CreateUser(db *sql.DB, username, password, qqID string) error {
	query := `INSERT INTO 用户 (username, password, qq_id) VALUES (?, ?, ?)`
	_, err := db.Exec(query, username, password, qqID)
	return err
}

// CreateSession creates a new session for a user with a default title
func CreateSession(db *sql.DB, username string) (int64, error) {
	defaultTitle := "新会话"
	query := `INSERT INTO 会话聊天 (username, title) VALUES (?, ?)`
	result, err := db.Exec(query, username, defaultTitle)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// RenameSession renames an existing session
func RenameSession(db *sql.DB, sessionID int, newTitle string) error {
	query := `UPDATE 会话聊天 SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, newTitle, sessionID)
	return err
}

// GetAllSessions retrieves all session data for a specific user
func GetAllSessions(db *sql.DB, username string) ([]map[string]interface{}, error) {
	query := `
	SELECT id, title, created_at, updated_at
	FROM 会话聊天
	WHERE username = ?
	ORDER BY created_at ASC
	`
	rows, err := db.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []map[string]interface{}
	for rows.Next() {
		var id int
		var title string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &title, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		session := map[string]interface{}{
			"id":         id,
			"title":      title,
			"created_at": createdAt,
			"updated_at": updatedAt,
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// InsertMessage adds a new message to a session
func InsertMessage(db *sql.DB, conversationID int, role string, senderName string, content string, tokenCount int) (int64, error) {
	query := `INSERT INTO 消息 (conversation_id, role, senderName,content, token_count) VALUES (?, ?, ?, ? ,?)`
	result, err := db.Exec(query, conversationID, role, senderName, content, tokenCount)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetSessionMessages retrieves the latest n messages from a session
func GetSessionMessages(db *sql.DB, conversationID, limit int) ([]map[string]interface{}, error) {
	query := `
	SELECT id, role, senderName, created_at, token_count, content
	FROM (
		SELECT id, role, senderName, created_at, token_count, content
		FROM 消息
		WHERE conversation_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	) AS sub
	ORDER BY created_at ASC
`
	rows, err := db.Query(query, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var id, tokenCount int
		var role, content string
		var senderName sql.NullString
		var createdAt time.Time

		err := rows.Scan(&id, &role, &senderName, &createdAt, &tokenCount, &content)
		if err != nil {
			return nil, err
		}

		message := map[string]interface{}{
			"id":          id,
			"role":        role,
			"created_at":  createdAt,
			"token_count": tokenCount,
			"content":     content,
			"senderName":  senderName.String,
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// CreateEmptyMessage 创建一条内容为空的消息并返回其ID
func CreateEmptyMessage(db *sql.DB, conversationID int, role string, senderName string) (int64, error) {
	// 创建一条内容为空，token数为0的消息
	query := `INSERT INTO 消息 (conversation_id, role, senderName, content, token_count) 
              VALUES (?, ?, ?, '', 0)`
	result, err := db.Exec(query, conversationID, role, senderName)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateMessageContent 更新指定ID消息的内容和token数
func UpdateMessageContent(db *sql.DB, messageID int64, content string, tokenCount int) error {
	query := `UPDATE 消息 
              SET content = ?, token_count = ?, created_at = CURRENT_TIMESTAMP 
              WHERE id = ?`
	_, err := db.Exec(query, content, tokenCount, messageID)
	return err
}
