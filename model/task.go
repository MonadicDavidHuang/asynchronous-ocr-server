package model

import "time"

const (
	TASK_STATUS_PENDING  = "pending"
	TASK_STATUS_COMPLETE = "complete"
	TASK_STATUS_DELETED  = "deleted"

	IMAGE_FILE_STATUS_UPLOADED = "uploaded"
	IMAGE_FILE_STATUS_DELETED  = "deleted"
)

type Task struct {
	ID              int64     `json:"id"`
	OpenTaskID      string    `json:"open_task_id"`
	TaskStatus      string    `json:"task_status"`
	ImageFileID     *int64    `json:"image_file_id"` // nullable
	ImageFileStatus string    `json:"image_file_status"`
	Caption         *string   `json:"caption"` // nullable
	UpdatedAt       time.Time `json:"updated_at"`
	CreatedAt       time.Time `json:"created_at"`
	// Status soft_delete.DeletedAt `gorm:"softDelete:flag"`
}

func (Task) TableName() string {
	return "tasks"
}
