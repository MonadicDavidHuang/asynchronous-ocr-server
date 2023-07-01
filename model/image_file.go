package model

import "time"

const (
	IMAGE_FILE_FILE_TYPE_JPEG = "image/jpeg"
	IMAGE_FILE_FILE_TYPE_PNG  = "image/png"
	IMAGE_FILE_FILE_TYPE_TIFF = "image/tiff"
)

type ImageFile struct {
	ID        int64     `json:"id"`
	Content   []byte    `json:"content" gorm:"type:blob"`
	FileType  string    `json:"file_type" gorm:"type:blob"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
	// Status soft_delete.DeletedAt `gorm:"softDelete:flag"`
}

func (ImageFile) TableName() string {
	return "image_files"
}
