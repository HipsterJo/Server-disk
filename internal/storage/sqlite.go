package storage

import (
	"database/sql"
	"disk-server/internal/lib/api/response"
	entities "disk-server/internal/lib/entities"
	"fmt"
	"log"
	"net/http"
	"time"

	"disk-server/internal/lib/conversion"

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

func GetUserByUserName(s *Storage, username string) (entities.UserDocument, response.Response) {
	const op = "storage.postgres.GetUserByUserName"

	var user entities.UserDocument

	row := s.db.QueryRow(`SELECT id, username, password, email FROM users WHERE username = $1`, username)

	err := row.Scan(&user.Id, &user.Username, &user.Password, &user.Email)
	if err != nil {
		return entities.UserDocument{}, response.Error("Пользователь не найден!")
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

func InitialFilter(r *http.Request) entities.FileQueryParams {
	var queryParams entities.FileQueryParams
	queryParams.Search = r.URL.Query().Get("search")
	minSizeStr := r.URL.Query().Get("minsize")
	maxSizeStr := r.URL.Query().Get("maxsize")
	queryParams.MimeType = r.URL.Query().Get("MimeType")

	queryParams.MinSize = conversion.StringToInt(minSizeStr, -1)
	queryParams.MaxSize = conversion.StringToInt(maxSizeStr, -1)

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	startDate, err := time.Parse("20060102", startDateStr)
	endDate, err := time.Parse("20060102", endDateStr)
	if err != nil {
	}

	queryParams.StartDate = startDate
	queryParams.EndDate = endDate

	return queryParams
}

func GetUserFiles(db *Storage, user entities.UserDocument, params entities.FileQueryParams, r *http.Request) ([]entities.File, response.Response) {
	sqlQuery := fmt.Sprintf(`
        SELECT filename, size, mime_type, uploaded_at, updated_at, path
        FROM files
        WHERE user_id='%d'
    `, user.Id)
	if params.Search != "" {
		sqlQuery += fmt.Sprintf(` AND filename LIKE '%%%s%%'`, params.Search)
	}
	if params.MinSize > 0 {
		sqlQuery += fmt.Sprintf(` AND size >= %d`, params.MinSize)
	}
	if params.MaxSize > 0 {
		sqlQuery += fmt.Sprintf(` AND size <= %d`, params.MaxSize)
	}
	if params.MimeType != "" {
		sqlQuery += fmt.Sprintf(` AND mime_type = '%s'`, params.MimeType)
	}
	if params.StartDate != (time.Time{}) {
		sqlQuery += fmt.Sprintf(` AND uploaded_at >= '%s'`, params.StartDate.Format("20060102"))
	}
	if params.EndDate != (time.Time{}) {
		sqlQuery += fmt.Sprintf(` AND updated_at <= '%s'`, params.EndDate.Format("20060102"))
	}

	switch params.SortBy {
	case entities.SortByName:
		sqlQuery += " ORDER BY filename"
	case entities.SortBySize:
		sqlQuery += " ORDER BY size"
	case entities.SortByUpload:
		sqlQuery += " ORDER BY uploaded_at"
	case entities.SortByUpdated:
		sqlQuery += " ORDER BY updated_at"
	}

	if params.SortDir == entities.SortDesc {
		sqlQuery += " DESC"
	}

	fmt.Println(sqlQuery)

	rows, err := db.db.Query(sqlQuery)
	if err != nil {
		log.Println("Failed to execute SQL query:", err)
		return nil, response.Error("Failed to fetch files")
	}
	defer rows.Close()

	var files []entities.File
	for rows.Next() {
		var file entities.File
		if err := rows.Scan(&file.FileName, &file.Size, &file.Mime_type, &file.Upload_at, &file.Updated_at, &file.Path); err != nil {
			log.Println("Failed to scan row:", err)
			return nil, response.Error("Failed to fetch files")
		}
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error while iterating over rows:", err)
		return nil, response.Error("Failed to fetch files")
	}

	fmt.Println(files)
	return files, response.OKWithData(files)
}
