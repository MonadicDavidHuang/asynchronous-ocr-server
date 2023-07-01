package worker_test

import (
	"asynchronous-ocr-server/imageutil"
	"asynchronous-ocr-server/model"
	"asynchronous-ocr-server/mysql"
	"asynchronous-ocr-server/repository"
	"asynchronous-ocr-server/worker"
	"context"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestOCRWorker(t *testing.T) {
	db, dbForFixtures := mysql.PrepareDBForTest()

	sut, _, taskService, _, taskRepository, _, client := worker.Preperation(db)
	defer client.Close()

	content := imageutil.GetImageBlob("../sample/English-Class-Memes.jpg")
	openTaskID := ulid.Make().String()

	mysql.PrepareTestFixturesFor("Empty", dbForFixtures)

	task, internalErr := taskService.CreateTask(context.Background(), content, imageutil.IMAGE_TYPE_JPEG, openTaskID)
	if internalErr != nil {
		panic(internalErr)
	}

	err := sut.ApplyOCR(context.Background())
	assert.Nil(t, err)

	task, internalErr = taskRepository.Get(context.Background(), nil, model.Task{OpenTaskID: task.OpenTaskID}, nil)
	if internalErr != nil {
		panic(internalErr)
	}

	assert.Equal(t, model.TASK_STATUS_COMPLETE, task.TaskStatus)
	assert.Equal(t, model.IMAGE_FILE_STATUS_UPLOADED, task.ImageFileStatus)
	assert.NotNil(t, task.Caption)
}

func TestImageFileDeleteWorker(t *testing.T) {
	db, dbForFixtures := mysql.PrepareDBForTest()

	ocrWorker, sut, taskService, _, taskRepository, imageFileRepository, client := worker.Preperation(db)
	defer client.Close()

	content := imageutil.GetImageBlob("../sample/English-Class-Memes.jpg")
	openTaskID := ulid.Make().String()

	mysql.PrepareTestFixturesFor("Empty", dbForFixtures)

	task, internalErr := taskService.CreateTask(context.Background(), content, imageutil.IMAGE_TYPE_JPEG, openTaskID)
	if internalErr != nil {
		panic(internalErr)
	}

	ocrWorker.ApplyOCR(context.Background())

	task, internalErr = taskService.GetTask(context.Background(), task.OpenTaskID) // update task.TaskStatus to "deleted"
	if internalErr != nil {
		panic(internalErr)
	}

	err := sut.DeleteImageFile(context.Background())
	assert.Nil(t, err)

	task, internalErr = taskRepository.Get(context.Background(), nil, model.Task{OpenTaskID: task.OpenTaskID}, nil)
	if internalErr != nil {
		panic(internalErr)
	}

	imageFileID := *task.ImageFileID

	assert.Equal(t, model.TASK_STATUS_DELETED, task.TaskStatus)
	assert.Equal(t, model.IMAGE_FILE_STATUS_DELETED, task.ImageFileStatus)
	assert.NotNil(t, task.Caption)

	_, interalErr := imageFileRepository.GetByID(context.Background(), imageFileID)
	assert.NotNil(t, interalErr)
	assert.Equal(t, repository.InternalErrorCodeNoRecordFoundError, interalErr.InternalErrorCode())
}
