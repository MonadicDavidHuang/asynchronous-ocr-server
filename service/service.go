package service

import (
	ettot "asynchronous-ocr-server/error"
	"asynchronous-ocr-server/imageutil"
	"asynchronous-ocr-server/model"
	"asynchronous-ocr-server/repository"
	"context"
	"fmt"

	"github.com/otiai10/gosseract/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TaskService interface {
	GetTask(ctx context.Context, openTaskID string) (model.Task, ettot.Error)
	CreateTask(ctx context.Context, content []byte, fileType string, openTaskID string) (model.Task, ettot.Error)
}

type taskServiceImpl struct {
	db                  *gorm.DB
	taskRepository      repository.TaskRepository
	imageFileRepository repository.ImageFileRepository
}

func (ts taskServiceImpl) GetTask(ctx context.Context, openTaskID string) (model.Task, ettot.Error) {
	task, internalErr := ts.taskRepository.Get(
		ctx,
		nil,
		model.Task{
			OpenTaskID: openTaskID,
			TaskStatus: model.TASK_STATUS_PENDING,
		},
		[]model.Task{
			{
				OpenTaskID: openTaskID,
				TaskStatus: model.TASK_STATUS_COMPLETE,
			},
			{
				OpenTaskID: openTaskID,
				TaskStatus: model.TASK_STATUS_DELETED,
			},
		},
	)
	if internalErr != nil {
		if internalErr.InternalErrorCode() == repository.InternalErrorCodeNoRecordFoundError {
			log.WithContext(ctx).Info(internalErr)

			return model.Task{}, NewServiceError(internalErr, InternalErrorCodeNoTaskFoundError)
		} else {
			log.WithContext(ctx).Error(internalErr)

			return model.Task{}, NewServiceError(internalErr, InternalErrorCodeFailedToGetTaskError)
		}
	}

	switch task.TaskStatus {
	case model.TASK_STATUS_PENDING:
		err := fmt.Errorf("task is still pending, open_task_id: %s", task.OpenTaskID)
		log.WithContext(ctx).Info(err)
		return model.Task{}, NewServiceError(err, InternalErrorCodeTaskIsPendingError)
	case model.TASK_STATUS_COMPLETE:
		_, internalErr = ts.taskRepository.Update(
			ctx,
			nil,
			model.Task{
				ID:         task.ID,
				TaskStatus: model.TASK_STATUS_DELETED,
			},
		)
		if internalErr != nil {
			log.WithContext(ctx).Info(internalErr)
			return model.Task{}, NewServiceError(internalErr, InternalErrorCodeFailedToDeleteTaskError)
		}

		return task, nil
	case model.TASK_STATUS_DELETED:
		err := fmt.Errorf("task is already complete and user has already queried result open_task_id: %s", task.OpenTaskID)
		log.WithContext(ctx).Info("baka")
		return model.Task{}, NewServiceError(err, InternalErrorCodeTaskIsDeletedError)
	default:
		err := fmt.Errorf("task_status must be either %s or %s, but it's %s, open_task_id: %s", model.TASK_STATUS_PENDING, model.TASK_STATUS_COMPLETE, task.TaskStatus, task.OpenTaskID)
		log.WithContext(ctx).Error(err)
		return model.Task{}, NewServiceError(err, InternalErrorCodeFailedToGetTaskError)
	}
}

func (ts taskServiceImpl) CreateTask(ctx context.Context, content []byte, fileType string, openTaskID string) (model.Task, ettot.Error) {
	imageFile, internalErr := ts.imageFileRepository.Create(
		ctx,
		model.ImageFile{
			Content:  content,
			FileType: fileType,
		},
	)
	if internalErr != nil {
		log.WithContext(ctx).Error(internalErr)
		return model.Task{}, NewServiceError(internalErr, InternalErrorCodeFailedToStoreImageFileError)
	}

	task, internalErr := ts.taskRepository.Create(
		ctx,
		model.Task{
			OpenTaskID:      openTaskID,
			TaskStatus:      model.TASK_STATUS_PENDING,
			ImageFileStatus: model.IMAGE_FILE_STATUS_UPLOADED,
			ImageFileID:     &imageFile.ID,
		})
	if internalErr != nil {
		log.WithContext(ctx).Error(internalErr)
		return model.Task{}, NewServiceError(internalErr, InternalErrorCodeFailedToCreateTaskError)
	}

	return task, nil
}

func NewTaskServiceImpl(
	db *gorm.DB,
	taskRepository repository.TaskRepository,
	imageFileRepository repository.ImageFileRepository,
) TaskService {
	return taskServiceImpl{
		db:                  db,
		taskRepository:      taskRepository,
		imageFileRepository: imageFileRepository,
	}
}

type OCRService interface {
	ApplyOCR(ctx context.Context, content []byte, fileType string) (string, ettot.Error)
}

type ocrServiceImpl struct {
	client *gosseract.Client
}

func (ocrs ocrServiceImpl) ApplyOCR(ctx context.Context, content []byte, fileType string) (string, ettot.Error) {
	filePath, internalErr := imageutil.SaveBlobAsImageFile(ctx, content, "/tmp", fileType)
	if internalErr != nil {
		log.WithContext(ctx).Error(internalErr)
		return "", internalErr

	}

	err := ocrs.client.SetImage(filePath)
	if err != nil {
		log.WithContext(ctx).Error(err)
		return "", NewServiceError(err, InternalErrorCodeFailedToApplyOCRError)
	}

	text, err := ocrs.client.Text()
	if err != nil {
		log.WithContext(ctx).Error(err)
		return "", NewServiceError(err, InternalErrorCodeFailedToApplyOCRError)
	}

	return text, nil
}

func NewOCRServiceImpl(
	client *gosseract.Client,
) OCRService {
	return ocrServiceImpl{
		client: client,
	}
}
