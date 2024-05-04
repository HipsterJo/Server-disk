package storage

import (
	"database/sql"
	"disk-server/internal/lib/api/response"
	entities "disk-server/internal/lib/entities"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(connStr string) (*Storage, error) {
	const op = "storage.postgres.NewStorage"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func CreateUser(u entities.User, s *Storage) error {
	const op = "storage.postgres.CreateUser"

	_, err := s.db.Exec(`
        INSERT INTO users(username, password, email) VALUES($1, $2, $3)
    `, u.Username, u.Password, u.Email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func GetAllUsers(s *Storage) ([]entities.User, error) {
	const op = "storage.postgres.GetAllUsers"

	rows, err := s.db.Query("SELECT username, password, email FROM users")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []entities.User

	for rows.Next() {
		var user entities.User

		if err := rows.Scan(&user.Username, &user.Password, &user.Email); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

func GetUserByUserName(s *Storage, username string) (entities.User, response.Response) {
	const op = "storage.postgres.GetUserByUserName"

	var user entities.User

	row := s.db.QueryRow(`
        SELECT username, password, email FROM users WHERE username = $1
    `, username)

	err := row.Scan(&user.Username, &user.Password, &user.Email)
	if err != nil {
		return entities.User{}, response.Error("Пользователь не найден!")
	}

	return user, response.OK()
}

func NewFile(s *Storage, username string, file entities.FileData) (entities.File, response.Response) {
	const op = "storage.postger.NewFile"

	row := s.db.QueryRow(`
		SELECT id FROM users WHERE username = $1
	`, username)
	var user_id int

	err := row.Scan(&user_id)
	if err != nil {
		return entities.File{}, response.Error("Пользователь не найден!")
	}

	stmt, err := s.db.Prepare(`
		INSERT INTO files (user_id, filename, size, mime_type , path)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, uploaded_at, updated_at
	`)
	if err != nil {
		fmt.Println(err)
		return entities.File{}, response.Error(fmt.Sprintf("%s: %s", op, err.Error()))
	}
	defer stmt.Close()

	var fileID string
	var uploaded_at time.Time
	var updated_at time.Time

	err = stmt.QueryRow(user_id, file.FileName, file.Size, file.Mime_type, file.Path).Scan(&fileID, &uploaded_at, &updated_at)
	if err != nil {
		return entities.File{}, response.Error(fmt.Sprintf("%s: %s", op, err.Error()))
	}

	newFile := entities.File{
		User_id:    user_id,
		Id:         fileID,
		FileName:   file.FileName,
		Size:       file.Size,
		Mime_type:  file.Mime_type,
		Path:       file.Path,
		Updated_at: updated_at,
		Upload_at:  uploaded_at,
	}

	return newFile, response.OKWithData(newFile)
}
