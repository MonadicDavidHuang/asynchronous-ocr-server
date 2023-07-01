package main

import (
	log "github.com/sirupsen/logrus"

	"asynchronous-ocr-server/config"
	"asynchronous-ocr-server/handler"
	"asynchronous-ocr-server/mysql"
	"asynchronous-ocr-server/repository"
	"asynchronous-ocr-server/service"
	"asynchronous-ocr-server/worker"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/otiai10/gosseract/v2"
)

func main() {
	log.SetReportCaller(true)

	isForTest := false
	config := config.Init(isForTest)
	db := mysql.Init(config, isForTest)

	gosseractClient := gosseract.NewClient()
	defer gosseractClient.Close()

	taskRepository := repository.NewTaskRepositoryImpl(db)
	imageFileRepository := repository.NewImageFileRepositoryImpl(db)

	taskService := service.NewTaskServiceImpl(db, taskRepository, imageFileRepository)
	ocrService := service.NewOCRServiceImpl(gosseractClient)

	newTaskSubmissionNotifier := make(chan int8, 100)
	newTaskDeletionNotifier := make(chan int8, 100)

	ocrWorker := worker.NewOCRWorker(db, taskRepository, imageFileRepository, ocrService)
	imageFileDeleteWorker := worker.NewImageFileDeleteWorker(db, taskRepository, imageFileRepository)

	worker.StartWorkers(
		ocrWorker,
		imageFileDeleteWorker,
		newTaskSubmissionNotifier,
		newTaskDeletionNotifier,
	)

	asyncOCRHandler := handler.NewAsyncOCRHandler(
		taskService,
		ocrService,
		newTaskSubmissionNotifier,
		newTaskDeletionNotifier,
	)

	e := newRouter(asyncOCRHandler)

	e.Logger.Fatal(e.Start(":1323"))
}

func newRouter(asyncOCRHandler handler.AsyncOCRHandler) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	e.POST("/image-sync", asyncOCRHandler.ApplyOCRImmediately)
	e.POST("/image", asyncOCRHandler.SubmitOCRTask)
	e.GET("/image", asyncOCRHandler.CheckOCRTask)

	return e
}
