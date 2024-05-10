package entities

import (
	"time"
)

type EnumSort int

const (
	SortByName EnumSort = iota
	SortBySize
	SortByUpload
	SortByUpdated
)

type EnumDir int

const (
	SortAsc EnumDir = iota
	SortDesc
)

type File struct {
	UserId      int       `json: "user_id"`
	Id          string    `json:"id"`
	FileName    string    `json: "filename"`
	Size        int       `json: "size`
	MimeType    string    `json: "mime_type"`
	UploadAt    time.Time `json: "uploaded_at"`
	UpdatedAt   time.Time `json: "updated_at"`
	Path        string    `json:"path"`
	FolderId    int       `json:"folder_id"`
	AccessLevel int       `json:"access_level"`
}

type FileData struct {
	FileName    string `json: "filename"`
	Size        int    `json: "size`
	MimeType    string `json: "mime_type"`
	Path        string `json:"path"`
	FolderId    int    `json: "folder_id"`
	AccessLevel int
}

type FileQueryParams struct {
	Search      string
	MinSize     int
	MaxSize     int
	MimeType    string
	SortBy      EnumSort
	SortDir     EnumDir
	StartDate   time.Time
	EndDate     time.Time
	AccessLevel int
}

type Folder struct {
	Id             int
	Name           string
	ParentFolderId int
	UserId         int
}
