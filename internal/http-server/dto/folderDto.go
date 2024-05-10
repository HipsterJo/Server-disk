package dto

type CreateFolder struct {
	Name           string `json:"name"`
	ParentFolderId int    `json:"parent_folder_id"`
}

type FileToFolder struct {
	FildeId  int `json:"file_id"`
	FolderId int `json:"folder_id"`
}
