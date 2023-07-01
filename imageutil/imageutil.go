package imageutil

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"image/png"
	"os"

	ettot "asynchronous-ocr-server/error"
	"asynchronous-ocr-server/model"

	"github.com/gabriel-vasile/mimetype"
	"github.com/oklog/ulid/v2"
	log "github.com/sirupsen/logrus"
	"golang.org/x/image/tiff"
)

const (
	IMAGE_TYPE_JPEG = "image/jpeg"
	IMAGE_TYPE_PNG  = "image/png"
	IMAGE_TYPE_TIFF = "image/tiff"

	// [50001, 59999] are reserved for imageutil
	internalErrorCodeBase                       ettot.InternalErrorCode = 50000
	InternalErrorCodeNotSupportedImageTypeError ettot.InternalErrorCode = iota + 1 + internalErrorCodeBase
	InternalErrorCodeFailedToSaveImageError
)

var errorMap = map[ettot.InternalErrorCode]string{
	InternalErrorCodeNotSupportedImageTypeError: "only .jpeg and .png are supported, child-error(%s)",
	InternalErrorCodeFailedToSaveImageError:     "failed to save image on disk, child-error(%w)",
}

func SaveBlobAsImageFile(ctx context.Context, blob []byte, directoryPath string, fileTypeName string) (string, ettot.Error) {
	temporaryFileName := ulid.Make()

	var filePath string
	var err error

	switch fileTypeName {
	case model.IMAGE_FILE_FILE_TYPE_JPEG:
		filePath = fmt.Sprintf("%s/%s.%s", directoryPath, temporaryFileName, "jpg")
		err = saveBlobAsJpeg(blob, filePath)
	case model.IMAGE_FILE_FILE_TYPE_PNG:
		filePath = fmt.Sprintf("%s/%s.%s", directoryPath, temporaryFileName, "png")
		err = saveBlobAsPng(blob, filePath)
	case model.IMAGE_FILE_FILE_TYPE_TIFF:
		filePath = fmt.Sprintf("%s/%s.%s", directoryPath, temporaryFileName, "jpg")
		err = saveBlobAsTiff(blob, filePath)
	default:
		err = fmt.Errorf("file type is %s", fileTypeName)
		return "", NewImageUtilError(err, InternalErrorCodeNotSupportedImageTypeError)
	}

	if err != nil {
		log.WithContext(ctx).Error(err)

		return "", NewImageUtilError(err, InternalErrorCodeFailedToSaveImageError)
	}

	return filePath, nil
}

func saveBlobAsJpeg(blob []byte, filePath string) error {
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}

	img, err := jpeg.Decode(bytes.NewReader(blob))
	if err != nil {
		return err
	}

	err = jpeg.Encode(out, img, nil)
	if err != nil {
		return err
	}

	return nil
}

func saveBlobAsPng(blob []byte, filePath string) error {
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}

	img, err := png.Decode(bytes.NewReader(blob))
	if err != nil {
		return err
	}

	err = png.Encode(out, img)
	if err != nil {
		return err
	}

	return nil
}

func saveBlobAsTiff(blob []byte, filePath string) error {
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}

	img, err := tiff.Decode(bytes.NewReader(blob))
	if err != nil {
		return err
	}

	err = tiff.Encode(out, img, nil)
	if err != nil {
		return err
	}

	return nil
}

func GetImageType(blob []byte) (string, ettot.Error) {
	mtype := mimetype.Detect(blob)

	if mtype.Is(model.IMAGE_FILE_FILE_TYPE_JPEG) {
		return model.IMAGE_FILE_FILE_TYPE_JPEG, nil
	}

	if mtype.Is(model.IMAGE_FILE_FILE_TYPE_PNG) {
		return model.IMAGE_FILE_FILE_TYPE_PNG, nil
	}

	if mtype.Is(model.IMAGE_FILE_FILE_TYPE_TIFF) {
		return model.IMAGE_FILE_FILE_TYPE_TIFF, nil
	}

	err := fmt.Errorf("file type is %s", mtype.String())

	return "", NewImageUtilError(err, InternalErrorCodeNotSupportedImageTypeError)
}

type ImageUtilError struct {
	Err               error
	internalErrorCode ettot.InternalErrorCode
}

func (e *ImageUtilError) InternalErrorCode() ettot.InternalErrorCode {
	return e.internalErrorCode
}

func (e *ImageUtilError) Error() string {
	msg := errorMap[e.internalErrorCode]
	return fmt.Errorf(msg, e.Err).Error()
}

func (e *ImageUtilError) Unwrap() error {
	return e.Err
}

func NewImageUtilError(
	err error,
	internalErrorCode ettot.InternalErrorCode,
) ettot.Error {
	return &ImageUtilError{
		err,
		internalErrorCode,
	}
}

// Test utility functions

func GetImageBlob(imagePath string) []byte {
	dat, err := os.ReadFile(imagePath)
	if err != nil {
		panic(err)
	}

	return dat
}
