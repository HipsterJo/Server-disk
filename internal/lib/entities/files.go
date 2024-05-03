package entities

import (
	"time"
)

type File struct {
	User_id    string    `json: "user_id"`
	Id         string    `json:"id"`
	FileName   string    `json: "filename"`
	Size       int       `json: "size`
	Mime_type  string    `json: "mime_type"`
	Upload_at  time.Time `json: "uploaded_at"`
	Updated_at time.Time `json: "updated_at"`
	Path       string    `json:"path"`
}
