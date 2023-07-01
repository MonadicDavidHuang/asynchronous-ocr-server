package imageutil_test

import (
	"asynchronous-ocr-server/imageutil"
	"asynchronous-ocr-server/model"
	"context"
	"encoding/base64"
	_ "image/gif"
	_ "image/png"

	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveBlobAsImageFile(t *testing.T) {
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

		blobOriginal := imageutil.GetImageBlob(imagePath)
		base64EncodedImage := base64.StdEncoding.EncodeToString(blobOriginal)

		blobDecoded, err := base64.StdEncoding.DecodeString(base64EncodedImage)
		if err != nil {
			panic(err)
		}

		savedImagePath, internalErr := imageutil.SaveBlobAsImageFile(context.Background(), blobDecoded, "../sample", imageType)
		assert.Nil(t, internalErr)

		// clean-up
		err = os.Remove(savedImagePath)
		if err != nil {
			panic(err)
		}

		assert.Equal(t, blobOriginal, blobDecoded)
	}

	_, internalErr := imageutil.SaveBlobAsImageFile(context.Background(), []byte{}, "../sample", "image/gif")
	assert.NotNil(t, internalErr)
	assert.Equal(t, imageutil.InternalErrorCodeNotSupportedImageTypeError, internalErr.InternalErrorCode())
}

func TestGetImageType_Supported(t *testing.T) {
	imagePath := "../sample/sample.tiff"

	blobOriginal := imageutil.GetImageBlob(imagePath)

	imageType, internalErr := imageutil.GetImageType(blobOriginal)
	assert.Nil(t, internalErr)

	assert.Equal(t, model.IMAGE_FILE_FILE_TYPE_TIFF, imageType)
}

func TestGetImageType_NotSupported(t *testing.T) {
	imagePath := "../sample/sample.gif"

	blobOriginal := imageutil.GetImageBlob(imagePath)

	_, internalErr := imageutil.GetImageType(blobOriginal)
	assert.NotNil(t, internalErr)
}
