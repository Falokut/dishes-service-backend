package entity

type UploadFileRequest struct {
	Category string `validate:"required"`
	Filename string
	Content  []byte `validate:"required"`
}

type FileRequest struct {
	Filename string `validate:"required"`
	Category string `validate:"required"`
}
