PROJECT?=github.com/donskova1ex/AverageRegionIncomes
READER_NAME?=excel_reader
READER_VERSION?=0.0.1
READER_CONTAINER_NAME?=docker.io/donskova1ex/${READER_NAME}

reader_local_build:
	go build -o bin/${READER_NAME} cmd/readers/${READER_NAME}.go

reader_docker_build:
	docker build -t ${READER_CONTAINER_NAME}:${READER_VERSION} -t ${READER_CONTAINER_NAME}:latest -f Dockerfile.reader .
