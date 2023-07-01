package mysql

import (
	"asynchronous-ocr-server/config"
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/go-testfixtures/testfixtures/v3"
	log "github.com/sirupsen/logrus"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

var (
	db   *gorm.DB
	once sync.Once
)

const (
	DevBox = "devbox"
)

// Init ..., call Init in your main func.
func Init(config config.Config, isForTest bool) *gorm.DB {
	once.Do(func() {
		db = constructDB(config, isForTest)
	})

	return db
}

func constructDB(config config.Config, isForTest bool) *gorm.DB {
	user := config.User
	pass := config.Pass
	host := config.Host
	port := config.Port
	database := config.Database

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user,
		pass,
		host,
		port,
		database,
	)

	var gormConfig gorm.Config
	if !isForTest {
		gormConfig = gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		}
	}

	containerizedDB, err := gorm.Open(mysql.Open(dsn), &gormConfig)
	if err != nil {
		log.Fatal(err)

	}

	msg := fmt.Sprintf("dsn: %s", dsn)
	log.Info(msg)

	if isForTest {
		containerizedDB = containerizedDB.Debug()
	}

	return containerizedDB
}

// WriteDB ...
func WriteDB(ctx context.Context) *gorm.DB {
	return db.Clauses(dbresolver.Write).WithContext(ctx)
}

// ReadDB ...
func ReadDB(ctx context.Context) *gorm.DB {
	return db.Clauses(dbresolver.Read).WithContext(ctx)
}

// DB ..., read write separation
func DB(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}

// Test utility functions

func PrepareDBForTest() (*gorm.DB, *sql.DB) {
	config := config.Init(true)
	Init(config, true)

	db := ReadDB(context.Background())
	dbForFixtures, err := db.DB()
	if err != nil {
		panic(err)
	}

	return db, dbForFixtures
}

func PrepareTestFixturesFor(testScenarioName string, dbForFixtures *sql.DB) {
	directoryPath := fmt.Sprintf("./fixtures/%s", testScenarioName)

	loadTestFixtures(dbForFixtures, directoryPath)
}

func loadTestFixtures(sqlDB *sql.DB, directoryPath string) {
	fixtures, err := testfixtures.New(
		testfixtures.Database(sqlDB),          // You database connection
		testfixtures.Dialect("mysql"),         // Available: "postgresql", "timescaledb", "mysql", "mariadb", "sqlite" and "sqlserver"
		testfixtures.Directory(directoryPath), // The directory containing the YAML files
	)
	if err != nil {
		panic(err)
	}

	err = fixtures.Load()
	if err != nil {
		panic(err)
	}
}
