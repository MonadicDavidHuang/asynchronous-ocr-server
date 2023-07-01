# asynchronous-ocr-server
This repository is for the implementation of the asynchronous OCR server.


The `asynchronous-ocr-server` will server the following features.
- apply OCR immediately
- submit OCR task and get `task_id`
- check submitted task with `task_id`

Note that after the task is completed and once user has checked the result, the target image of OCR will be deleted from system's data-store,
and task itself will be marked as `deleted`.

## Dependency

### Go
The following command will install all the dependencies required for this repository.
```
asdf install
```

### golang-migrate
Also you need to install [golang-migrate](https://github.com/golang-migrate/migrate#cli-usage) to operate migration on MySQL.
```
go install -tags mysql github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Docker
As this project uses Docker for running MySQL for testing and demo purpose, please also [install Docker](https://docs.docker.com/engine/install/).

### tesseract
As this project uses tesseract for execute OCR, please also [install tesseract](https://tesseract-ocr.github.io/tessdoc/Installation.html).

Below version is confirmed to be working.
```
tesseract 5.3.1
 leptonica-1.82.0
  libgif 5.2.1 : libjpeg 8d (libjpeg-turbo 2.1.5.1) : libpng 1.6.40 : libtiff 4.5.1 : zlib 1.2.11 : libwebp 1.3.0 : libopenjp2 2.5.0
 Found AVX2
 Found AVX
 Found FMA
 Found SSE4.1
 Found libarchive 3.6.2 zlib/1.2.11 liblzma/5.4.1 bz2lib/1.0.8 liblz4/1.9.4 libzstd/1.5.4
 Found libcurl/7.79.1 SecureTransport (LibreSSL/3.3.6) zlib/1.2.11 nghttp2/1.45.1
```

## Build

The following command will generate binary: `asynchronous-ocr-server` on repository root, which is the binary you use to run server.

```
go build
```

## Run asynchronous-ocr-server
The `asynchronous-ocr-server` uses MySQL as database.
The MySQL database must be prepared and migrated with proper schemas BEFORE the `asynchronous-ocr-server` server runs.

### Prepare Database
#### Using prepared Database
If you have real MySQL, please make new file under `config/configs/config_${environment_name}.yml`,
and put required information as the following.
```
# config_${environment_name}.yml
db_host: "127.0.0.1"
db_port: "3307"
db_database: "test-asynchronous-ocr-server-database"
db_user: "asynchronous-ocr-server"
db_pass: "asynchronous-ocr-server"
```

Also please note that to use that MySQL, you have to set environment variable as the following.
```
export PROFILE=${environment_name}
```

#### Using containerized Database
Run following command to run MySQL as docker-container.
```
docker-compose up -d
```

Once the MySQL is started, run following command to migrate the database.
```
migrate -database="mysql://asynchronous-ocr-server:asynchronous-ocr-server@tcp(127.0.0.1:3307)/test-asynchronous-ocr-server-database" -path=db/migrations/ up 2
```

Please note that the corresponding `${environment_name}` for above configuration is `local` which is set as default.

### Run Server
The following commands will launch `asynchronous-ocr-server` (using build binary).
```
./asynchronous-ocr-server
```

You can also launch server with the following command.
```
go run ./main.go
```

## Running Test
As this project contains several tests using actual database, 
and the testfixture could cause dead-lock in the data preperation phase, 
please run go test as the following.
```
go test -p 1 ./...
```

Note that you have to prepare local MySQL (with `docker-compose` command) and complete database migration to run those tests.

## Sample Usage
### POST /image-sync
Request:
```
curl --location 'http://localhost:1323/image-sync' \
--header 'Content-Type: application/json' \
--data '{
    "image_data": "<b64 encoded image>"
}'
```

Response:
```
{
    "text": "ME WHEN THE CLASSROOM\nBOOK ORDER ARRIVES\n7. <!\n\\ A Win! | ——\nied , { =~"
}
```

### POST /image
Request:
```
curl --location 'http://localhost:1323/image' \
--header 'Content-Type: application/json' \
--data '{
    "image_data": "<b64 encoded image>"
}'
```

Response:
```
{
    "task_id": "01H48Z0WMWXVH6TY33VEZ5DNVE"
}
```

### Get /image (task is pending)
Request:
```
curl --location --request GET 'http://localhost:1323/image' \
--header 'Content-Type: application/json' \
--data '{
    "task_id": "01H48Z43KCNCHHYDJ557DZJGCB"
}'
```

Response:
```
{
    "text": "null"
}
```

### Get /image (task is complete)
Request:
```
curl --location --request GET 'http://localhost:1323/image' \
--header 'Content-Type: application/json' \
--data '{
    "task_id": "01H48Z43KCNCHHYDJ557DZJGCB"
}'
```

Response:
```
{
    "text": "ME WHEN THE CLASSROOM\nBOOK ORDER ARRIVES\n7. <!\n\\ A Win! | ——\nied , { =~"
}
```


### Get /image (task is pending)
Request:
```
curl --location --request GET 'http://localhost:1323/image' \
--header 'Content-Type: application/json' \
--data '{
    "task_id": "01H48Z43KCNCHHYDJ557DZJGCB"
}'
```

Response:
```
{
    "text": "null"
}
```

## Verified OS information

```
# system_profiler SPSoftwareDataType SPHardwareDataType
Software:

    System Software Overview:

      System Version: macOS 12.6 (21G115)
      Kernel Version: Darwin 21.6.0
      Boot Volume: Macintosh HD
      Boot Mode: Normal
      Computer Name: C02CPTFJMD6M
      User Name: ByteDance (bytedance)
      Secure Virtual Memory: Enabled
      System Integrity Protection: Enabled
      Time since boot: 3 days 23:46

Hardware:

    Hardware Overview:

      Model Name: MacBook Pro
      Model Identifier: MacBookPro16,1
      Processor Name: 6-Core Intel Core i7
      Processor Speed: 2.6 GHz
      Number of Processors: 1
      Total Number of Cores: 6
      L2 Cache (per Core): 256 KB
      L3 Cache: 12 MB
      Hyper-Threading Technology: Enabled
      Memory: 16 GB
      System Firmware Version: 1731.140.2.0.0 (iBridge: 19.16.16067.0.0,0)
      OS Loader Version: 540.120.3~22
      Serial Number (system): C02CPTFJMD6M
      Hardware UUID: 34AD58F8-55AA-5788-A550-228CDB7C30B0
      Provisioning UDID: 34AD58F8-55AA-5788-A550-228CDB7C30B0
      Activation Lock Status: Disabled
```

## Remaining Tasks
- Add more unit-tests with un-happy path
- Add swagger for API
