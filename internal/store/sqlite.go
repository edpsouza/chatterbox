package store

import (
	"database/sql"
	"errors"

	"github.com/edpsouza/chatterbox/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

// Store wraps the SQLite DB connection.
type Store struct {
	db *sql.DB
}

// NewStore initializes the SQLite database and returns a Store.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		return nil, err
	}
	return store, nil
}

// migrate creates necessary tables if they don't exist.
func (s *Store) migrate() error {
	var err error
	var rows *sql.Rows
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		public_key TEXT,
		status TEXT DEFAULT 'offline',
		last_seen DATETIME
	);`
	_, err = s.db.Exec(userTable)
	if err != nil {
		return err
	}

	// Automated migration: add status and last_seen columns if missing
	rows, err = s.db.Query("PRAGMA table_info(users);")
	if err != nil {
		return err
	}
	columns := map[string]bool{}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		columns[name] = true
	}
	if !columns["status"] {
		_, err := s.db.Exec("ALTER TABLE users ADD COLUMN status TEXT DEFAULT 'offline'")
		if err != nil {
			return err
		}
	}
	if !columns["last_seen"] {
		_, err := s.db.Exec("ALTER TABLE users ADD COLUMN last_seen DATETIME")
		if err != nil {
			return err
		}
	}
	messageTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		username TEXT NOT NULL,
		recipient TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`
	_, err = s.db.Exec(messageTable)
	if err != nil {
		return err
	}

	// Automated migration: add recipient column if missing
	rows, err = s.db.Query("PRAGMA table_info(messages);")
	if err != nil {
		return err
	}
	columns = map[string]bool{}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		columns[name] = true
	}
	if !columns["recipient"] {
		_, err = s.db.Exec("ALTER TABLE messages ADD COLUMN recipient TEXT NOT NULL DEFAULT ''")
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateUser inserts a new user into the database.
func (s *Store) CreateUser(user *models.User) error {
	stmt := `INSERT INTO users (username, password, public_key) VALUES (?, ?, ?)`
	result, err := s.db.Exec(stmt, user.Username, user.Password, user.PublicKey)
	if err != nil {
		if sqliteIsUniqueConstraint(err) {
			return errors.New("username already exists")
		}
		return err
	}
	id, err := result.LastInsertId()
	if err == nil {
		user.ID = id
	}
	return nil
}

// GetUserByUsername fetches a user by username.
func (s *Store) GetUserByUsername(username string) (*models.User, error) {
	stmt := `SELECT id, username, password, public_key, status, last_seen FROM users WHERE LOWER(username) = LOWER(?)`

	row := s.db.QueryRow(stmt, username)
	var user models.User
	var lastSeen sql.NullString
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.PublicKey, &user.Status, &lastSeen)
	if lastSeen.Valid {
		user.LastSeen = lastSeen.String
	} else {
		user.LastSeen = ""
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// sqliteIsUniqueConstraint checks if an error is a SQLite unique constraint violation.
func sqliteIsUniqueConstraint(err error) bool {
	return err != nil && (err.Error() == "UNIQUE constraint failed: users.username" ||
		err.Error() == "UNIQUE constraint failed: user.username")
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// CreateMessage inserts a new chat message into the database.
func (s *Store) CreateMessage(userID int64, username, recipient, content string) error {
	stmt := `INSERT INTO messages (user_id, username, recipient, content) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(stmt, userID, username, recipient, content)
	return err
}

// GetRecentMessages fetches the most recent N messages.
func (s *Store) GetRecentMessages(limit int) ([]struct {
	ID        int64
	UserID    int64
	Username  string
	Content   string
	CreatedAt string
}, error) {
	stmt := `SELECT id, user_id, username, content, created_at FROM messages ORDER BY created_at DESC LIMIT ?`
	rows, err := s.db.Query(stmt, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []struct {
		ID        int64
		UserID    int64
		Username  string
		Content   string
		CreatedAt string
	}
	for rows.Next() {
		var m struct {
			ID        int64
			UserID    int64
			Username  string
			Content   string
			CreatedAt string
		}
		if err := rows.Scan(&m.ID, &m.UserID, &m.Username, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

// GetMessagesBetween fetches encrypted messages exchanged between two users, ordered by created_at ascending.
func (s *Store) GetMessagesBetween(userA, userB string) ([]models.Message, error) {
	stmt := `
		SELECT id, user_id, username, recipient, content, created_at
		FROM messages
		WHERE (username = ? AND recipient = ?)
		   OR (username = ? AND recipient = ?)
		ORDER BY created_at ASC
	`
	rows, err := s.db.Query(stmt, userA, userB, userB, userA)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.Username, &m.Recipient, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

// SetUserStatus updates the user's status (e.g., "online", "offline").
func (s *Store) SetUserStatus(username, status string) error {
	stmt := `UPDATE users SET status = ? WHERE username = ?`
	_, err := s.db.Exec(stmt, status, username)
	return err
}

// SetUserLastSeen updates the user's last_seen timestamp.
func (s *Store) SetUserLastSeen(username, timestamp string) error {
	stmt := `UPDATE users SET last_seen = ? WHERE username = ?`
	_, err := s.db.Exec(stmt, timestamp, username)
	return err
}

// SetUserLastSeenNow updates the user's last_seen timestamp to the current time.
func (s *Store) SetUserLastSeenNow(username string) error {
	stmt := `UPDATE users SET last_seen = CURRENT_TIMESTAMP WHERE username = ?`
	_, err := s.db.Exec(stmt, username)
	return err
}
