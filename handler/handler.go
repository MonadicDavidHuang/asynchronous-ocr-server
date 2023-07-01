package handler

import (
	"asynchronous-ocr-server/imageutil"
	"asynchronous-ocr-server/service"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/oklog/ulid/v2"
	log "github.com/sirupsen/logrus"

	"net/http"

	"github.com/labstack/echo/v4"
)

type AsyncOCRHandler struct {
	taskService               service.TaskService
	ocrService                service.OCRService
	newTaskSubmissionNotifier chan int8
	newTaskDeletionNotifier   chan int8
}

func NewAsyncOCRHandler(
	taskService service.TaskService,
	ocrService service.OCRService,
	newTaskSubmissionNotifier chan int8,
	newTaskDeletionNotifier chan int8,
) AsyncOCRHandler {
	return AsyncOCRHandler{
		taskService:               taskService,
		ocrService:                ocrService,
		newTaskSubmissionNotifier: newTaskSubmissionNotifier,
		newTaskDeletionNotifier:   newTaskDeletionNotifier,
	}
}

func (aocrh AsyncOCRHandler) ApplyOCRImmediately(c echo.Context) error {
	ctx := c.Request().Context()

	jsonBody, err := GetJSONRawBody(c)
	if err != nil {
		log.WithContext(ctx).Error(err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	base64EncodedBlob, found := jsonBody["image_data"]
	if !found {
		err = fmt.Errorf("request body did not include image_data")
		log.WithContext(ctx).Error(err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	content, err := base64.StdEncoding.DecodeString(base64EncodedBlob)
	if err != nil {
		log.WithContext(ctx).Error(err)
		return c.String(http.StatusBadRequest, "Failed to decode as base64 encoded image data")
	}

	fileType, internalErr := imageutil.GetImageType(content)
	if internalErr != nil {
		log.WithContext(ctx).Error(err)

		switch internalErr.InternalErrorCode() {
		case imageutil.InternalErrorCodeNotSupportedImageTypeError:
			return c.String(http.StatusBadRequest, "Passed image file's format is not supported")
		default:
			return c.String(http.StatusInternalServerError, "Something wrong happened")
		}
	}

	text, internalErr := aocrh.ocrService.ApplyOCR(ctx, content, fileType)
	if internalErr != nil {
		log.WithContext(ctx).Error(internalErr)

		switch internalErr.InternalErrorCode() {
		case service.InternalErrorCodeFailedToApplyOCRError:
			return c.String(http.StatusBadRequest, "Failed to apply ocr")
		case imageutil.InternalErrorCodeFailedToSaveImageError:
			return c.String(http.StatusBadRequest, "Failed to save image for applying ocr")
		case imageutil.InternalErrorCodeNotSupportedImageTypeError:
			return c.String(http.StatusBadRequest, "Passed image file's format is not supported")
		default:
			return c.String(http.StatusInternalServerError, "Something wrong happened")
		}
	}

	response := map[string]string{
		"text": text,
	}

	return c.JSON(http.StatusOK, response)
}

func (aocrh AsyncOCRHandler) SubmitOCRTask(c echo.Context) error {
	ctx := c.Request().Context()

	jsonBody, err := GetJSONRawBody(c)
	if err != nil {
		log.WithContext(ctx).Error(err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	base64EncodedBlob, found := jsonBody["image_data"]
	if !found {
		err = fmt.Errorf("request body did not include image_data")
		log.WithContext(ctx).Error(err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	content, err := base64.StdEncoding.DecodeString(base64EncodedBlob)
	if err != nil {
		log.WithContext(ctx).Error(err)
		return c.String(http.StatusBadRequest, "Failed to decode as base64 encoded image data")
	}

	fileType, internalErr := imageutil.GetImageType(content)
	if internalErr != nil {
		log.WithContext(ctx).Error(err)

		switch internalErr.InternalErrorCode() {
		case imageutil.InternalErrorCodeNotSupportedImageTypeError:
			return c.String(http.StatusBadRequest, "Passed image file's format is not supported")
		default:
			return c.String(http.StatusInternalServerError, "Something wrong happened")
		}
	}

	task, internalErr := aocrh.taskService.CreateTask(ctx, content, fileType, ulid.Make().String())
	if internalErr != nil {
		log.WithContext(ctx).Error(err)

		switch internalErr.InternalErrorCode() {
		case service.InternalErrorCodeFailedToStoreImageFileError:
			return c.String(http.StatusInternalServerError, "Could not store image into data store")
		case service.InternalErrorCodeFailedToCreateTaskError:
			return c.String(http.StatusInternalServerError, "Could not create task")
		default:
			return c.String(http.StatusInternalServerError, "Something wrong happened")
		}
	}

	defer func() {
		if aocrh.newTaskSubmissionNotifier != nil {
			aocrh.newTaskSubmissionNotifier <- int8(0)
		}
	}()

	response := map[string]string{
		"task_id": task.OpenTaskID,
	}

	return c.JSON(http.StatusOK, response)
}

func (aocrh AsyncOCRHandler) CheckOCRTask(c echo.Context) error {
	ctx := c.Request().Context()

	jsonBody, err := GetJSONRawBody(c)
	if err != nil {
		log.WithContext(ctx).Error(err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	key := "task_id"

	openTaskID, found := jsonBody[key]
	if !found {
		err = fmt.Errorf("request body did not include key: %s", key)
		log.WithContext(ctx).Error(err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	task, internalErr := aocrh.taskService.GetTask(ctx, openTaskID)
	if internalErr != nil {
		switch internalErr.InternalErrorCode() {
		case service.InternalErrorCodeNoTaskFoundError:
			log.WithContext(ctx).Info(internalErr)
			return c.String(http.StatusNotFound, "No queried task exist")
		case service.InternalErrorCodeTaskIsPendingError, service.InternalErrorCodeTaskIsDeletedError:
			log.WithContext(ctx).Info(internalErr)

			defer func() {
				if aocrh.newTaskDeletionNotifier != nil {
					aocrh.newTaskDeletionNotifier <- int8(0)
				}
			}()

			response := map[string]string{
				"text": "null",
			}

			return c.JSON(http.StatusOK, response)
		case service.InternalErrorCodeFailedToGetTaskError:
			log.WithContext(ctx).Error(internalErr)
			return c.String(http.StatusInternalServerError, "Failed to get queried task")
		case service.InternalErrorCodeFailedToDeleteTaskError:
			log.WithContext(ctx).Error(internalErr)
			return c.String(http.StatusInternalServerError, "Failed to delete queried task")
		default:
			log.WithContext(ctx).Error(internalErr)
			return c.String(http.StatusInternalServerError, "Something wrong happened")
		}
	}

	response := map[string]string{
		"text": *task.Caption,
	}

	return c.JSON(http.StatusOK, response)
}

func GetJSONRawBody(c echo.Context) (map[string]string, error) {
	jsonBody := make(map[string]string)

	err := json.NewDecoder(c.Request().Body).Decode(&jsonBody)
	if err != nil {
		return nil, err
	}

	return jsonBody, nil
}

// Test utility functions

func Serialize(key, text string) ([]byte, error) {
	response := map[string]string{
		key: text,
	}

	b, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		return nil, err
	}

	return b, nil
}
