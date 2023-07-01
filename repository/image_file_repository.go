package repository

import (
	ettot "asynchronous-ocr-server/error"
	"asynchronous-ocr-server/model"
	"context"

	"gorm.io/gorm"
)

//go:generate mockgen -source=image_file_repository.go -destination mock/mock_image_file_repository.go
type ImageFileRepository interface {
	GetByID(ctx context.Context, id int64) (model.ImageFile, ettot.Error)
	Create(ctx context.Context, toBeCreated model.ImageFile) (model.ImageFile, ettot.Error)
	DeleteByID(ctx context.Context, id int64) ettot.Error
}

type imageFileRepositoryImpl struct {
	db *gorm.DB
}

func (ifr imageFileRepositoryImpl) GetByID(ctx context.Context, id int64) (model.ImageFile, ettot.Error) {
	var imageFile model.ImageFile

	query := ifr.db.WithContext(ctx).Where(&model.ImageFile{ID: id})

	result := query.First(&imageFile)
	err := result.Error
	if err != nil {
		doProperLogging(ctx, err)
		return model.ImageFile{}, getProperError(err)
	}

	return imageFile, nil
}

func (ifr imageFileRepositoryImpl) Create(ctx context.Context, toBeCreated model.ImageFile) (model.ImageFile, ettot.Error) {
	result := ifr.db.Create(&toBeCreated)
	err := result.Error
	if err != nil {
		doProperLogging(ctx, err)
		return model.ImageFile{}, getProperError(err)
	}

	return toBeCreated, nil
}

func (ifr imageFileRepositoryImpl) DeleteByID(ctx context.Context, id int64) ettot.Error {
	result := ifr.db.WithContext(ctx).Delete(&model.ImageFile{ID: id})
	err := result.Error
	if err != nil {
		doProperLogging(ctx, err)
		return getProperError(err)
	}

	return nil
}

func NewImageFileRepositoryImpl(db *gorm.DB) ImageFileRepository {
	return imageFileRepositoryImpl{db: db}
}
