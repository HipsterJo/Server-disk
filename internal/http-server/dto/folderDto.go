package dto

type CreateFolder struct {
	Name           string `json:"name"`
	ParentFolderId int    `json:"parent_folder_id"`
}
