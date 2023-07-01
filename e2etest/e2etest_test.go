package e2etest_test

import (
	"asynchronous-ocr-server/handler"
	"asynchronous-ocr-server/imageutil"
	"asynchronous-ocr-server/model"
	"asynchronous-ocr-server/mysql"
	"asynchronous-ocr-server/repository"
	"asynchronous-ocr-server/worker"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

func TestServer_WithoutWorkers(t *testing.T) {
	// preparing resources for test
	db, dbForFixtures := mysql.PrepareDBForTest()

	ocrWorker, _, taskService, ocrService, _, _, gosseractClient := worker.Preperation(db)
	defer gosseractClient.Close()

	asyncOCRHandler := handler.NewAsyncOCRHandler(taskService, ocrService, nil, nil)

	mysql.PrepareTestFixturesFor("Empty", dbForFixtures)

	router := newRouter(asyncOCRHandler)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	client := new(http.Client)

	// POST /image-sync
	{
		req := constructPostImageRequest(testServer.URL+"/image-sync", "../sample/English-Class-Memes.jpg")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		response := map[string]string{}
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		text, found := response["text"]
		assert.True(t, found)
		assert.Greater(t, len(text), 0)
	}

	// GET /image, not found
	{
		req := constructGetImageRequest(testServer.URL+"/image", "122333444455555")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	}

	// POST /image
	openTaskID := func() string {
		req := constructPostImageRequest(testServer.URL+"/image", "../sample/English-Class-Memes.jpg")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		response := map[string]string{}
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		openTaskID, found := response["task_id"]
		assert.True(t, found)
		assert.Greater(t, len(openTaskID), 0)

		return openTaskID
	}()

	// GET /image, ok, but still pending
	{
		req := constructGetImageRequest(testServer.URL+"/image", openTaskID)

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		response := map[string]string{}
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		text, found := response["text"]
		assert.True(t, found)
		assert.Equal(t, "null", text)
	}

	// run OCR
	ocrWorker.ApplyOCR(context.Background())

	// GET /image, ok, get text
	{
		req := constructGetImageRequest(testServer.URL+"/image", openTaskID)

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		response := map[string]string{}
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		text, found := response["text"]
		assert.True(t, found)
		assert.Greater(t, len(text), 0)
	}

	// GET /image, ok, but alrady deleted
	{
		req := constructGetImageRequest(testServer.URL+"/image", openTaskID)

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		response := map[string]string{}
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		text, found := response["text"]
		assert.True(t, found)
		assert.Equal(t, "null", text)
	}
}

func TestServer_WithWorkers(t *testing.T) {
	// preparing resources for test
	db, dbForFixtures := mysql.PrepareDBForTest()

	ocrWorker, imageFileDeleteWorker, taskService, ocrService, taskRepository, imageFileRepository, gosseractClient := worker.Preperation(db)
	defer gosseractClient.Close()

	newTaskSubmissionNotifier := make(chan int8, 100)
	newTaskDeletionNotifier := make(chan int8, 100)

	worker.StartWorkers(ocrWorker, imageFileDeleteWorker, newTaskSubmissionNotifier, newTaskDeletionNotifier)

	asyncOCRHandler := handler.NewAsyncOCRHandler(taskService, ocrService, newTaskSubmissionNotifier, newTaskDeletionNotifier)

	mysql.PrepareTestFixturesFor("Empty", dbForFixtures)

	router := newRouter(asyncOCRHandler)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	client := new(http.Client)

	// POST /image-sync
	{
		req := constructPostImageRequest(testServer.URL+"/image-sync", "../sample/English-Class-Memes.jpg")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		response := map[string]string{}
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		text, found := response["text"]
		assert.True(t, found)
		assert.Greater(t, len(text), 0)
	}

	// GET /image, not found
	{
		req := constructGetImageRequest(testServer.URL+"/image", "122333444455555")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	}

	// POST /image
	openTaskID := func() string {
		req := constructPostImageRequest(testServer.URL+"/image", "../sample/English-Class-Memes.jpg")

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		response := map[string]string{}
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		openTaskID, found := response["task_id"]
		assert.True(t, found)
		assert.Greater(t, len(openTaskID), 0)

		return openTaskID
	}()

	time.Sleep(time.Second * 3) // sleep for waiting OCR worker

	// GET /image, ok, get text
	{
		req := constructGetImageRequest(testServer.URL+"/image", openTaskID)

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		response := map[string]string{}
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		text, found := response["text"]
		assert.True(t, found)
		assert.Greater(t, len(text), 0)
	}

	// GET /image, ok, but alrady deleted
	{
		req := constructGetImageRequest(testServer.URL+"/image", openTaskID)

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		response := map[string]string{}
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		text, found := response["text"]
		assert.True(t, found)
		assert.Equal(t, "null", text)
	}

	time.Sleep(time.Second * 3) // sleep for waiting image file delete worker

	task, internalErr := taskRepository.Get(context.Background(), nil, model.Task{OpenTaskID: openTaskID}, nil)
	if internalErr != nil {
		panic(internalErr)
	}
	assert.Equal(t, model.TASK_STATUS_DELETED, task.TaskStatus)
	assert.Equal(t, model.IMAGE_FILE_STATUS_DELETED, task.ImageFileStatus)
	assert.NotNil(t, task.ImageFileID)

	_, internalErr = imageFileRepository.GetByID(context.Background(), *task.ImageFileID)
	assert.Equal(t, repository.InternalErrorCodeNoRecordFoundError, internalErr.InternalErrorCode())

}

func constructPostImageRequest(uri, imagePath string) *http.Request {
	image := imageutil.GetImageBlob(imagePath)
	base64EncodedImage := base64.StdEncoding.EncodeToString(image)

	body, err := handler.Serialize("image_data", base64EncodedImage)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", uri, bytes.NewReader(body))
	if err != nil {
		panic(err)
	}

	return req
}

func constructGetImageRequest(uri, openTaskID string) *http.Request {
	body, err := handler.Serialize("task_id", openTaskID)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("GET", uri, bytes.NewReader(body))
	if err != nil {
		panic(err)
	}

	return req
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
