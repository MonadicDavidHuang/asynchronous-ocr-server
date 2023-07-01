package mysql_test

import (
	"asynchronous-ocr-server/model"
	"asynchronous-ocr-server/mysql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDB_VerifyLocalDB(t *testing.T) {
	db, dbForFixtures := mysql.PrepareDBForTest()

	mysql.PrepareTestFixturesFor("Sample", dbForFixtures)

	var count int64

	err := db.Model(&model.Task{}).Count(&count).Error
	assert.Nil(t, err)
	assert.Equal(t, int64(4), count)

	err = db.Model(&model.ImageFile{}).Count(&count).Error
	assert.Nil(t, err)
	assert.Equal(t, int64(4), count)
}
