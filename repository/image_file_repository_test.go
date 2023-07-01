package repository_test

import (
	"asynchronous-ocr-server/model"
	"asynchronous-ocr-server/mysql"
	"asynchronous-ocr-server/repository"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageFileRepositoryImpl_GetByID(t *testing.T) {
	db, dbForFixtures := mysql.PrepareDBForTest()

	mysql.PrepareTestFixturesFor("TestImageFileRepository_GetByID", dbForFixtures)

	sut := repository.NewImageFileRepositoryImpl(db)

	imageFileID := int64(1)
	content := []byte("test")
	fileType := model.IMAGE_FILE_FILE_TYPE_JPEG

	imageFile, internalErr := sut.GetByID(context.Background(), imageFileID)
	assert.Nil(t, internalErr)

	assert.Equal(t, content, imageFile.Content)
	assert.Equal(t, fileType, imageFile.FileType)
}

func TestImageFileRepositoryImpl_Create(t *testing.T) {
	db, dbForFixtures := mysql.PrepareDBForTest()

	mysql.PrepareTestFixturesFor("Empty", dbForFixtures)

	sut := repository.NewImageFileRepositoryImpl(db)

	content := []byte("test")
	fileType := model.IMAGE_FILE_FILE_TYPE_JPEG

	imageFile, internalErr := sut.Create(
		context.Background(),
		model.ImageFile{
			Content:  content,
			FileType: fileType,
		},
	)
	assert.Nil(t, internalErr)

	imageFile, internalErr = sut.GetByID(context.Background(), imageFile.ID)
	assert.Nil(t, internalErr)

	assert.Equal(t, content, imageFile.Content)
	assert.Equal(t, fileType, imageFile.FileType)
}

func TestImageFileRepositoryImpl_Delete(t *testing.T) {
	db, dbForFixtures := mysql.PrepareDBForTest()

	mysql.PrepareTestFixturesFor("TestImageFileRepository_Delete", dbForFixtures)

	sut := repository.NewImageFileRepositoryImpl(db)

	imageFileID := int64(1)

	internalErr := sut.DeleteByID(context.Background(), imageFileID)
	assert.Nil(t, internalErr)

	_, internalErr = sut.GetByID(context.Background(), imageFileID)
	assert.Equal(t, repository.InternalErrorCodeNoRecordFoundError, internalErr.InternalErrorCode())
}
