package storage

import (
	"database/sql"
	entities "disk-server/internal/lib/entities"
	"fmt"

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

	// Выполняем SQL-запрос для выборки всех пользователей
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

func GetUserByUserName(s *Storage, username string) (entities.User, error) {
	const op = "storage.postgres.GetUserByUserName"

	var user entities.User

	row := s.db.QueryRow(`
        SELECT username, password, email FROM users WHERE username = $1
    `, username)

	err := row.Scan(&user.Username, &user.Password, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.User{}, fmt.Errorf("%s: user not found", op)
		}
		return entities.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
