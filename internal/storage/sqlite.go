package storage

import (
	"database/sql"
	"disk-server/internal/http-server/dto"
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

	err = stmt.QueryRow(user_id, file.FileName, file.Size, file.MimeType, file.Path).Scan(&fileID, &uploaded_at, &updated_at)
	if err != nil {
		return entities.File{}, response.Error(fmt.Sprintf("%s: %s", op, err.Error()))
	}

	newFile := entities.File{
		UserId:    user_id,
		Id:        fileID,
		FileName:  file.FileName,
		Size:      file.Size,
		MimeType:  file.MimeType,
		Path:      file.Path,
		UpdatedAt: updated_at,
		UploadAt:  uploaded_at,
	}

	return newFile, response.OKWithData(newFile)
}

// Вынести в отдельный файл!!!!
func InitialFilter(r *http.Request) entities.FileQueryParams {
	var queryParams entities.FileQueryParams
	queryParams.Search = r.URL.Query().Get("search")
	accessLevelStr := r.URL.Query().Get("access_level")
	minSizeStr := r.URL.Query().Get("min_size")
	maxSizeStr := r.URL.Query().Get("max_size")
	queryParams.MimeType = r.URL.Query().Get("mime_type")

	queryParams.MinSize = conversion.StringToInt(minSizeStr, -1)
	queryParams.MaxSize = conversion.StringToInt(maxSizeStr, -1)
	queryParams.AccessLevel = conversion.StringToInt(accessLevelStr, 0)

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

// Надо отправлять файлы "с рабочего пространства" и массив с папками? Или сделать отдельный рут initialFiles?
func GetUserFiles(db *Storage, user entities.UserDocument, params entities.FileQueryParams, r *http.Request) ([]entities.File, response.Response) {
	sqlQuery := fmt.Sprintf(`
        SELECT id, filename, size, mime_type, uploaded_at, updated_at, path, access_level
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
		if err := rows.Scan(&file.Id, &file.FileName, &file.Size, &file.MimeType, &file.UploadAt, &file.UpdatedAt, &file.Path, &file.AccessLevel); err != nil {
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

func GetHomeDir(db *Storage, user entities.UserDocument, r *http.Request) response.Response {
	sqlQuery := fmt.Sprintf(`
        SELECT id, filename, size, mime_type, uploaded_at, updated_at, path, access_level
        FROM files
        WHERE user_id='%d' AND folder_id IS NULL
    `, user.Id)

	rows, err := db.db.Query(sqlQuery)
	if err != nil {
		return response.Error("Ошибка получения данных1")
	}
	defer rows.Close()

	var files []entities.File
	for rows.Next() {
		var file entities.File
		if err := rows.Scan(&file.Id, &file.FileName, &file.Size, &file.MimeType, &file.UploadAt, &file.UpdatedAt, &file.Path, &file.AccessLevel); err != nil {
			return response.Error("Ошибка получения данных")
		}
		fmt.Println(file)
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return response.Error("Ошибка получения данных")
	}

	sqlQuery = fmt.Sprintf(`
        SELECT id, name, user_id
        FROM folders
        WHERE user_id='%d' AND parent_folder_id IS NULL
    `, user.Id)

	rows, err = db.db.Query(sqlQuery)
	if err != nil {
		log.Println("Failed to execute SQL query:", err)
		return response.Error("Ошибка получения данных")
	}
	defer rows.Close()

	var folders []entities.Folder
	for rows.Next() {
		var folder entities.Folder
		if err := rows.Scan(&folder.Id, &folder.Name, &folder.UserId); err != nil {
			log.Println("Failed to scan row:", err)
			return response.Error("Ошибка получения данных")
		}
		folders = append(folders, folder)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error while iterating over rows:", err)
		return response.Error("Ошибка получения данных")
	}

	return response.OKWithData(map[string]interface{}{
		"files":   files,
		"folders": folders,
	})
}

func AddFileToFolder(db *Storage, folder_id int, file_id int, user entities.UserDocument) response.Response {

	sqlQuery := `
		SELECT id FROM folders WHERE id = $1 AND  user_id = $2
	`

	err := db.db.QueryRow(sqlQuery, folder_id, user.Id).Scan(&folder_id)
	if err != nil {
		log.Println("Failed to check if folder exists:", err)
		return response.Error("Папка не найдена!")
	}

	sqlQuery = `
		UPDATE files
		SET folder_id = $1
		WHERE id = $2 AND user_id = $3
	`
	_, err = db.db.Exec(sqlQuery, folder_id, file_id, user.Id)
	if err != nil {
		return response.Error("Ошибка при добавления файла в папку!")
	}

	return response.OK()

}

func CreateFolder(db *Storage, folderDto dto.CreateFolder, user entities.UserDocument) response.Response {

	if folderDto.ParentFolderId == 0 {
		sqlQuery := `
		INSERT INTO folders (name,  user_id)
		VALUES ($1, $2 )
		RETURNING id, name, user_id
	`
		row := db.db.QueryRow(sqlQuery, folderDto.Name, user.Id)
		var folder entities.Folder

		err := row.Scan(&folder.Id, &folder.Name, &folder.UserId)
		if err != nil {
			return response.Error(fmt.Sprintf("Failed to create folder: %v", err))
		}

		return response.OKWithData(folder)
	}

	sqlQuery := `
		INSERT INTO folders (name, parent_folder_id, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, name, parent_folder_id, user_id
	`

	row := db.db.QueryRow(sqlQuery, folderDto.Name, folderDto.ParentFolderId, user.Id)

	var folder entities.Folder

	err := row.Scan(&folder.Id, &folder.Name, &folder.ParentFolderId, &folder.UserId)
	if err != nil {
		return response.Error(fmt.Sprintf("Failed to create folder: %v", err))
	}

	return response.OKWithData(folder)
}

func GetFilesFromFolder(db *Storage, user entities.UserDocument, r *http.Request, folder_id int) response.Response {
	sqlQuery := fmt.Sprintf(`
        SELECT id, filename, size, mime_type, uploaded_at, updated_at, path, access_level
        FROM files
        WHERE user_id='%d' AND folder_id = '%d'
    `, user.Id, folder_id)

	rows, err := db.db.Query(sqlQuery)
	if err != nil {
		return response.Error("Ошибка получения данных")
	}
	defer rows.Close()

	var files []entities.File
	for rows.Next() {
		var file entities.File
		if err := rows.Scan(&file.Id, &file.FileName, &file.Size, &file.MimeType, &file.UploadAt, &file.UpdatedAt, &file.Path, &file.AccessLevel); err != nil {
			return response.Error("Ошибка получения данных")
		}
		fmt.Println(file)
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return response.Error("Ошибка получения данных")
	}

	sqlQuery = fmt.Sprintf(`
        SELECT id, name, user_id
        FROM folders
        WHERE user_id='%d' AND parent_folder_id='%d'
    `, user.Id, folder_id)

	rows, err = db.db.Query(sqlQuery)
	if err != nil {
		log.Println("Failed to execute SQL query:", err)
		return response.Error("Ошибка получения данных")
	}
	defer rows.Close()

	var folders []entities.Folder
	for rows.Next() {
		var folder entities.Folder
		if err := rows.Scan(&folder.Id, &folder.Name, &folder.UserId); err != nil {
			log.Println("Failed to scan row:", err)
			return response.Error("Ошибка получения данных")
		}
		folders = append(folders, folder)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error while iterating over rows:", err)
		return response.Error("Ошибка получения данных")
	}

	return response.OKWithData(map[string]interface{}{
		"files":   files,
		"folders": folders,
	})
}
