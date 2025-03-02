package context

import (
	"database/sql"
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
