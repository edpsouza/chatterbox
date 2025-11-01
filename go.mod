module github.com/edpsouza/chatterbox

go 1.21

require (
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/gorilla/websocket v1.5.0
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-sqlite3 v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.17.0
)

require golang.org/x/sys v0.15.0 // indirect

replace github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v1.14.17
