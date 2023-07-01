package repository_test

import (
	"asynchronous-ocr-server/model"
	"asynchronous-ocr-server/mysql"
	"asynchronous-ocr-server/repository"
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskRepositoryImpl_Get(t *testing.T) {
	db, dbForFixtures := mysql.PrepareDBForTest()

	sut := repository.NewTaskRepositoryImpl(db)

	id := int64(1)
	openTaskID := "01ARZ3NDEKTSV4RRFFQ69G5FAA"

	// by id, without transaction pattern
	mysql.PrepareTestFixturesFor("TestTaskRepository_Get", dbForFixtures)

	task, internalErr := sut.Get(context.Background(), nil, model.Task{ID: id}, nil)
	assert.Nil(t, internalErr)
	assert.Equal(t, openTaskID, task.OpenTaskID)

	// by id, with transaction pattern
	mysql.PrepareTestFixturesFor("TestTaskRepository_Get", dbForFixtures)

	tx := db.Begin()
	{
		task, internalErr := sut.Get(context.Background(), tx, model.Task{ID: id}, nil)
		assert.Nil(t, internalErr)
		assert.Equal(t, openTaskID, task.OpenTaskID)
	}
	tx.Commit()

	// by open_task_id, without transaction pattern
	mysql.PrepareTestFixturesFor("TestTaskRepository_Get", dbForFixtures)

	task, internalErr = sut.Get(context.Background(), nil, model.Task{OpenTaskID: openTaskID}, nil)
	assert.Nil(t, internalErr)
	assert.Equal(t, id, task.ID)

	// by open_task_id, with transaction pattern
	mysql.PrepareTestFixturesFor("TestTaskRepository_Get", dbForFixtures)

	tx = db.Begin()
	{
		task, internalErr := sut.Get(context.Background(), tx, model.Task{OpenTaskID: openTaskID}, nil)
		assert.Nil(t, internalErr)
		assert.Equal(t, id, task.ID)
	}
	tx.Commit()

	// by open_task_id and task_status, without transaction pattern, with optional specifiers
	mysql.PrepareTestFixturesFor("TestTaskRepository_Get", dbForFixtures)

	task, internalErr = sut.Get(
		context.Background(),
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
		},
	)
	assert.Nil(t, internalErr)
	assert.Equal(t, id, task.ID)
	assert.Equal(t, model.TASK_STATUS_PENDING, task.TaskStatus)

	// by open_task_id and task_status, with transaction pattern, with optional specifiers
	mysql.PrepareTestFixturesFor("TestTaskRepository_Get", dbForFixtures)

	tx = db.Begin()
	{
		task, internalErr = sut.Get(
			context.Background(),
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
			},
		)
		assert.Nil(t, internalErr)
		assert.Equal(t, id, task.ID)
		assert.Equal(t, model.TASK_STATUS_PENDING, task.TaskStatus)
	}
	tx.Commit()
}

func TestTaskRepositoryImpl_TryToGetOneWithLock(t *testing.T) {
	db, dbForFixtures := mysql.PrepareDBForTest()

	sut := repository.NewTaskRepositoryImpl(db)

	mysql.PrepareTestFixturesFor("TestTaskRepository_TryToGetOneWithLock", dbForFixtures)

	var wg sync.WaitGroup
	ch := make(chan int64, 9)

	for i := 0; i < 9; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			tx := db.Begin()
			{
				task, internalErr := sut.TryToGetOneWithLock(context.Background(), tx, model.Task{TaskStatus: model.TASK_STATUS_PENDING})
				if internalErr != nil {
					assert.Equal(t, repository.InternalErrorCodeNoRecordFoundError, internalErr.InternalErrorCode())
				}

				time.Sleep(time.Second * 1)

				ch <- task.ID
			}
			tx.Commit()
		}()
	}

	wg.Wait()

	close(ch)

	ids := []int64{}
	for id := range ch {
		ids = append(ids, id)
	}

	sort.SliceStable(ids, func(i, j int) bool { return ids[i] < ids[j] })

	expectedIDs := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8}

	assert.Equal(t, expectedIDs, ids)
}

func TestTaskRepositoryImpl_Create(t *testing.T) {
	db, dbForFixtures := mysql.PrepareDBForTest()

	mysql.PrepareTestFixturesFor("Empty", dbForFixtures)

	sut := repository.NewTaskRepositoryImpl(db)

	openTaskID := "01ARZ3NDEKTSV4RRFFQ69G5FAA"
	taskStatus := model.TASK_STATUS_PENDING
	imageFileStatus := model.IMAGE_FILE_STATUS_UPLOADED
	task := model.Task{
		OpenTaskID:      openTaskID,
		TaskStatus:      taskStatus,
		ImageFileID:     nil,
		ImageFileStatus: imageFileStatus,
	}

	task, internalErr := sut.Create(context.Background(), task)
	assert.Nil(t, internalErr)
	assert.Greater(t, task.ID, int64(-1))

	task, internalErr = sut.Get(context.Background(), nil, model.Task{ID: task.ID}, nil)
	assert.Nil(t, internalErr)
	assert.Equal(t, openTaskID, task.OpenTaskID)
	assert.Equal(t, taskStatus, task.TaskStatus)
	assert.Equal(t, imageFileStatus, task.ImageFileStatus)
}

func TestTaskRepositoryImpl_Update(t *testing.T) {
	db, dbForFixtures := mysql.PrepareDBForTest()

	sut := repository.NewTaskRepositoryImpl(db)

	id := int64(1)
	taskStatus := model.TASK_STATUS_COMPLETE
	imageFileStatus := model.IMAGE_FILE_STATUS_DELETED
	task := model.Task{
		ID:              id,
		TaskStatus:      taskStatus,
		ImageFileStatus: imageFileStatus,
	}

	// without transaction pattern
	mysql.PrepareTestFixturesFor("TestTaskRepository_Update", dbForFixtures)

	task, internalErr := sut.Update(context.Background(), nil, task)
	assert.Nil(t, internalErr)
	assert.Equal(t, taskStatus, task.TaskStatus)
	assert.Equal(t, imageFileStatus, task.ImageFileStatus)

	task, internalErr = sut.Get(context.Background(), nil, model.Task{ID: id}, nil)
	assert.Nil(t, internalErr)
	assert.Equal(t, taskStatus, task.TaskStatus)
	assert.Equal(t, imageFileStatus, task.ImageFileStatus)

	// with transaction pattern
	mysql.PrepareTestFixturesFor("TestTaskRepository_Update", dbForFixtures)

	tx := db.Begin()
	{
		task, internalErr := sut.Update(context.Background(), tx, task)
		assert.Nil(t, internalErr)
		assert.Equal(t, taskStatus, task.TaskStatus)
		assert.Equal(t, imageFileStatus, task.ImageFileStatus)
	}
	tx.Commit()

	task, internalErr = sut.Get(context.Background(), nil, model.Task{ID: id}, nil)
	assert.Nil(t, internalErr)
	assert.Equal(t, taskStatus, task.TaskStatus)
	assert.Equal(t, imageFileStatus, task.ImageFileStatus)
}
