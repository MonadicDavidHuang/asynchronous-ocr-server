package service_test

import (
	"asynchronous-ocr-server/imageutil"
	"asynchronous-ocr-server/model"
	mock_repository "asynchronous-ocr-server/repository/mock"

	"asynchronous-ocr-server/service"

	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/otiai10/gosseract/v2"
	"github.com/stretchr/testify/assert"
)

func TestTaskServiceImpl_GetTask(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	taskRepository := mock_repository.NewMockTaskRepository(mockCtrl)
	imageFileRepository := mock_repository.NewMockImageFileRepository(mockCtrl)

	ctx := context.Background()
	id := int64(1)
	openTaskID := "01ARZ3NDEKTSV4RRFFQ69G5FAA"
	task := model.Task{ID: id, OpenTaskID: openTaskID, TaskStatus: model.TASK_STATUS_COMPLETE}

	taskRepository.EXPECT().Get(
		gomock.Eq(ctx),
		nil,
		gomock.Eq(
			model.Task{OpenTaskID: openTaskID, TaskStatus: model.TASK_STATUS_PENDING},
		),
		gomock.Eq(
			[]model.Task{
				{OpenTaskID: openTaskID, TaskStatus: model.TASK_STATUS_COMPLETE},
				{OpenTaskID: openTaskID, TaskStatus: model.TASK_STATUS_DELETED},
			},
		),
	).Return(task, nil).Times(1)

	taskRepository.EXPECT().Update(
		gomock.Eq(ctx),
		nil,
		gomock.Eq(
			model.Task{ID: id, TaskStatus: model.TASK_STATUS_DELETED},
		),
	).Return(task, nil).Times(1)

	sut := service.NewTaskServiceImpl(nil, taskRepository, imageFileRepository)

	returnedApp, internalErr := sut.GetTask(ctx, openTaskID)
	assert.Nil(t, internalErr)

	assert.Equal(t, task, returnedApp)
}

func TestTaskServiceImpl_CreateTask(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	taskRepository := mock_repository.NewMockTaskRepository(mockCtrl)
	imageFileRepository := mock_repository.NewMockImageFileRepository(mockCtrl)

	ctx := context.Background()

	imageFileID := int64(1)
	content := []byte("test")
	fileType := model.IMAGE_FILE_FILE_TYPE_JPEG
	imageFile := model.ImageFile{ID: imageFileID, Content: content, FileType: fileType}

	taskID := int64(1)
	openTaskID := "01ARZ3NDEKTSV4RRFFQ69G5FAA"
	task := model.Task{ID: taskID, OpenTaskID: openTaskID, TaskStatus: model.TASK_STATUS_COMPLETE}

	imageFileRepository.EXPECT().Create(
		gomock.Eq(ctx),
		gomock.Eq(
			model.ImageFile{Content: content, FileType: fileType},
		),
	).Return(imageFile, nil).Times(1)

	taskRepository.EXPECT().Create(
		gomock.Eq(ctx),
		gomock.Eq(
			model.Task{
				OpenTaskID:      openTaskID,
				TaskStatus:      model.TASK_STATUS_PENDING,
				ImageFileStatus: model.IMAGE_FILE_STATUS_UPLOADED,
				ImageFileID:     &imageFileID,
			},
		),
	).Return(task, nil).Times(1)

	sut := service.NewTaskServiceImpl(nil, taskRepository, imageFileRepository)

	returnedApp, internalErr := sut.CreateTask(ctx, content, fileType, openTaskID)
	assert.Nil(t, internalErr)

	assert.Equal(t, task, returnedApp)
}

func TestOCRServiceImpl_ApplyOCR(t *testing.T) {
	client := gosseract.NewClient()
	defer client.Close()

	ocrService := service.NewOCRServiceImpl(client)

	imagePaths := []string{
		"../sample/sample.jpg",
		"../sample/sample.png",
		"../sample/sample.tiff",
	}

	imageTypes := []string{
		model.IMAGE_FILE_FILE_TYPE_JPEG,
		model.IMAGE_FILE_FILE_TYPE_PNG,
		model.IMAGE_FILE_FILE_TYPE_TIFF,
	}

	for i := 0; i < len(imagePaths); i++ {
		imagePath := imagePaths[i]
		imageType := imageTypes[i]

		content := imageutil.GetImageBlob(imagePath)

		_, internalErr := ocrService.ApplyOCR(context.Background(), content, imageType)
		assert.Nil(t, internalErr)
	}
}
