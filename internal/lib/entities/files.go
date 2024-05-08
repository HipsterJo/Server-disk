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
	User_id    int       `json: "user_id"`
	Id         string    `json:"id"`
	FileName   string    `json: "filename"`
	Size       int       `json: "size`
	Mime_type  string    `json: "mime_type"`
	Upload_at  time.Time `json: "uploaded_at"`
	Updated_at time.Time `json: "updated_at"`
	Path       string    `json:"path"`
}

type FileData struct {
	FileName  string `json: "filename"`
	Size      int    `json: "size`
	Mime_type string `json: "mime_type"`
	Path      string `json:"path"`
}

type FileQueryParams struct {
	Search    string
	MinSize   int
	MaxSize   int
	MimeType  string
	SortBy    EnumSort
	SortDir   EnumDir
	StartDate time.Time
	EndDate   time.Time
}
