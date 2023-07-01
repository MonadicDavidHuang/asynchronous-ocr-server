package repository

import (
	ettot "asynchronous-ocr-server/error"
	"context"
	"errors"

	log "github.com/sirupsen/logrus"

	"gorm.io/gorm"
)

func getProperError(err error) ettot.Error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewNRepositoryError(err, InternalErrorCodeNoRecordFoundError)
	}
	return NewNRepositoryError(err, InternalErrorCodeSystemRelatedError)
}

func doProperLogging(ctx context.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.WithContext(ctx).Info(err)
	} else {
		log.WithContext(ctx).Error(err)
	}
}
