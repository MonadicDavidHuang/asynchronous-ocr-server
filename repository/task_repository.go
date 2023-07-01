package repository

import (
	"asynchronous-ocr-server/model"
	"context"
	"errors"

	ettot "asynchronous-ocr-server/error"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNoTxIsPassed = errors.New("no transaction receiver is passed")
)

//go:generate mockgen -source=task_repository.go -destination mock/mock_task_repository.go
type TaskRepository interface {
	Get(ctx context.Context, tx *gorm.DB, specifier model.Task, optionalSpecifiers []model.Task) (model.Task, ettot.Error)
	TryToGetOneWithLock(ctx context.Context, tx *gorm.DB, specifier model.Task) (model.Task, ettot.Error)
	Create(ctx context.Context, toBeCreated model.Task) (model.Task, ettot.Error)
	Update(ctx context.Context, tx *gorm.DB, newTask model.Task) (model.Task, ettot.Error)
}

type taskRepositoryImpl struct {
	db *gorm.DB
}

func (tr taskRepositoryImpl) Get(
	ctx context.Context,
	tx *gorm.DB,
	specifier model.Task,
	optionalSpecifiers []model.Task,
) (model.Task, ettot.Error) {
	var query *gorm.DB

	if tx == nil {
		query = tr.db.WithContext(ctx)
	} else {
		query = tx.
			WithContext(ctx).
			Clauses(
				clause.Locking{
					Strength: "UPDATE",
				},
			)
	}

	var task model.Task

	query = query.Where(specifier)

	for _, optionalSpecifier := range optionalSpecifiers {
		query = query.Or(optionalSpecifier)
	}

	result := query.First(&task) // if it's during transaction, take lock
	err := result.Error
	if err != nil {
		doProperLogging(ctx, err)
		return model.Task{}, getProperError(err)
	}

	return task, nil
}

func (tr taskRepositoryImpl) TryToGetOneWithLock(ctx context.Context, tx *gorm.DB, specifier model.Task) (model.Task, ettot.Error) {
	if tx == nil {
		err := ErrNoTxIsPassed
		doProperLogging(ctx, err)
		return model.Task{}, getProperError(err)
	}

	var task model.Task

	query := tx.
		WithContext(ctx).
		Clauses(
			clause.Locking{
				Strength: "UPDATE",
				Options:  "SKIP LOCKED",
			},
		).
		Where(&specifier)

	result := query.First(&task) // take lock
	err := result.Error
	if err != nil {
		doProperLogging(ctx, err)
		return model.Task{}, getProperError(err)
	}

	return task, nil
}

func (tr taskRepositoryImpl) Create(
	ctx context.Context,
	toBeCreated model.Task,
) (model.Task, ettot.Error) {
	result := tr.db.WithContext(ctx).Create(&toBeCreated)
	err := result.Error
	if err != nil {
		doProperLogging(ctx, err)
		return model.Task{}, getProperError(err)
	}

	return toBeCreated, nil
}

func (tr taskRepositoryImpl) Update(
	ctx context.Context,
	tx *gorm.DB,
	newTask model.Task,
) (model.Task, ettot.Error) {
	if tx == nil {
		tx = tr.db
	}

	result := tx.WithContext(ctx).Updates(&newTask)
	err := result.Error
	if err != nil {
		doProperLogging(ctx, err)
		return model.Task{}, getProperError(err)
	}

	return newTask, nil
}

func NewTaskRepositoryImpl(db *gorm.DB) TaskRepository {
	return taskRepositoryImpl{db: db}
}
