package worker

import (
	"asynchronous-ocr-server/model"
	"asynchronous-ocr-server/repository"
	"asynchronous-ocr-server/service"
	"context"
	"fmt"

	"github.com/otiai10/gosseract/v2"
	log "github.com/sirupsen/logrus"

	"gorm.io/gorm"
)

func StartWorkers(
	ocrWorker OCRWorker,
	imagefileDeleteWOrker ImageFileDeleteWorker,
	newTaskSubmissionNotifier chan int8,
	newTaskDeletionNotifier chan int8,
) {
	for i := 0; i < 10; i++ {
		go func() {
			for range newTaskSubmissionNotifier {
				ocrWorker.ApplyOCR(context.Background())
			}
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			for range newTaskDeletionNotifier {
				imagefileDeleteWOrker.DeleteImageFile(context.Background())
			}
		}()
	}
}

type OCRWorker struct {
	db                  *gorm.DB
	taskRepository      repository.TaskRepository
	imageFileRepository repository.ImageFileRepository
	ocrService          service.OCRService
}

func (worker OCRWorker) ApplyOCR(ctx context.Context) error {
	var taskID int64
	var imageFileID int64
	var previousTaskStatus string
	var postTaskStatus string

	tx := worker.db.Begin()
	{
		task, internalErr := worker.taskRepository.TryToGetOneWithLock(ctx, tx, model.Task{TaskStatus: model.TASK_STATUS_PENDING, ImageFileStatus: model.IMAGE_FILE_STATUS_UPLOADED})
		if internalErr != nil {
			if internalErr.InternalErrorCode() != repository.InternalErrorCodeNoRecordFoundError {
				log.WithContext(ctx).Error(internalErr)
			}

			tx.Rollback()

			return internalErr // TODO: wrap with proper error
		}

		taskID = task.ID
		previousTaskStatus = task.TaskStatus

		if task.ImageFileID == nil {
			err := fmt.Errorf("task.ImageFileStatus is %s, but task.ImageFileID is nil, task.ID: %d", task.ImageFileStatus, task.ID)

			log.WithContext(ctx).Error(err)

			tx.Rollback()

			return err // TODO: wrap with proper error
		}

		imageFileID = *task.ImageFileID

		imageFile, internalErr := worker.imageFileRepository.GetByID(ctx, imageFileID)
		if internalErr != nil {
			log.WithContext(ctx).Error(internalErr)

			tx.Rollback()

			return internalErr // TODO: wrap with proper error
		}

		content := imageFile.Content
		fileType := imageFile.FileType

		text, internalErr := worker.ocrService.ApplyOCR(ctx, content, fileType)
		if internalErr != nil {
			log.WithContext(ctx).Error(internalErr)

			tx.Rollback()

			return internalErr // TODO: wrap with proper error
		}

		task, internalErr = worker.taskRepository.Update(ctx, tx, model.Task{ID: taskID, TaskStatus: model.TASK_STATUS_COMPLETE, Caption: &text})
		if internalErr != nil {
			log.WithContext(ctx).Error(internalErr)

			tx.Rollback()

			return internalErr // TODO: wrap with proper error
		}

		postTaskStatus = task.TaskStatus
	}
	err := tx.Commit().Error
	if err != nil {
		log.WithContext(ctx).Error(err)

		tx.Rollback()

		return err // TODO: wrap with proper error
	} else {
		log.WithContext(ctx).Info(
			fmt.Sprintf("OCR is applied for imageFile, and the task's task_status is updated from %s to %s, task.ID: %d, imageFile.ID: %d", previousTaskStatus, postTaskStatus, taskID, imageFileID),
		)
	}

	return nil
}

func NewOCRWorker(
	db *gorm.DB,
	taskRepository repository.TaskRepository,
	imageFileRepository repository.ImageFileRepository,
	ocrService service.OCRService,
) OCRWorker {
	return OCRWorker{
		db:                  db,
		taskRepository:      taskRepository,
		imageFileRepository: imageFileRepository,
		ocrService:          ocrService,
	}
}

type ImageFileDeleteWorker struct {
	db                  *gorm.DB
	taskRepository      repository.TaskRepository
	imageFileRepository repository.ImageFileRepository
}

func (worker ImageFileDeleteWorker) DeleteImageFile(ctx context.Context) error {
	var taskID int64
	var imageFileID int64
	var previousImageFileStatus string
	var postImageFileStatus string

	tx := worker.db.Begin()
	{
		task, internalErr := worker.taskRepository.TryToGetOneWithLock(ctx, tx, model.Task{TaskStatus: model.TASK_STATUS_DELETED, ImageFileStatus: model.IMAGE_FILE_STATUS_UPLOADED})
		if internalErr != nil {
			if internalErr.InternalErrorCode() != repository.InternalErrorCodeNoRecordFoundError {
				log.WithContext(ctx).Error(internalErr)
			}

			tx.Rollback()

			return internalErr // TODO: wrap with proper error
		}

		taskID = task.ID
		previousImageFileStatus = task.ImageFileStatus

		if task.ImageFileID == nil {
			err := fmt.Errorf("task.ImageFileStatus is %s, but task.ImageFileID is nil, task.ID: %d", task.ImageFileStatus, task.ID)

			log.WithContext(ctx).Error(err)

			tx.Rollback()

			return err // TODO: wrap with proper error
		}

		imageFileID = *task.ImageFileID

		internalErr = worker.imageFileRepository.DeleteByID(ctx, imageFileID)
		if internalErr != nil {
			log.WithContext(ctx).Error(internalErr)

			tx.Rollback()

			return internalErr // TODO: wrap with proper error
		}

		task, internalErr = worker.taskRepository.Update(ctx, tx, model.Task{ID: taskID, ImageFileStatus: model.IMAGE_FILE_STATUS_DELETED})
		if internalErr != nil {
			log.WithContext(ctx).Error(internalErr)

			tx.Rollback()

			return internalErr // TODO: wrap with proper error
		}

		postImageFileStatus = task.ImageFileStatus
	}
	err := tx.Commit().Error
	if err != nil {
		log.WithContext(ctx).Error(err)

		tx.Rollback()

		return err // TODO: wrap with proper error
	} else {
		log.WithContext(ctx).Info(
			fmt.Sprintf("imageFile for corresponding task is deleted and its image_file_status is updated from %s to %s, task.ID: %d, imageFile.ID: %d", previousImageFileStatus, postImageFileStatus, taskID, imageFileID),
		)
	}

	return nil
}

func NewImageFileDeleteWorker(
	db *gorm.DB,
	taskRepository repository.TaskRepository,
	imageFileRepository repository.ImageFileRepository,
) ImageFileDeleteWorker {
	return ImageFileDeleteWorker{
		db:                  db,
		taskRepository:      taskRepository,
		imageFileRepository: imageFileRepository,
	}
}

// Test utility methods

func Preperation(db *gorm.DB) (
	OCRWorker,
	ImageFileDeleteWorker,
	service.TaskService,
	service.OCRService,
	repository.TaskRepository,
	repository.ImageFileRepository,
	*gosseract.Client,
) {
	taskRepository := repository.NewTaskRepositoryImpl(db)
	imageFileRepository := repository.NewImageFileRepositoryImpl(db)

	taskService := service.NewTaskServiceImpl(db, taskRepository, imageFileRepository)

	gosserectClient := gosseract.NewClient()
	ocrService := service.NewOCRServiceImpl(gosserectClient)

	ocrWorker := NewOCRWorker(db, taskRepository, imageFileRepository, ocrService)
	imageFileDeleteWorker := NewImageFileDeleteWorker(db, taskRepository, imageFileRepository)

	return ocrWorker, imageFileDeleteWorker, taskService, ocrService, taskRepository, imageFileRepository, gosserectClient
}
