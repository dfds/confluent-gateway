APP_IMAGE_NAME=selfservice/confluent-gateway
DB_IMAGE_NAME=selfservice/confluent-gateway/dbmigrations
BUILD_NUMBER=n/a
OUTPUT_DIR=${PWD}/.output
OUTPUT_DIR_APP=${OUTPUT_DIR}/app
OUTPUT_DIR_MANIFESTS=${OUTPUT_DIR}/manifests
OUTPUT_DIR_TESTS=${OUTPUT_DIR}/tests
GOOS=linux
GOARCH=amd64
RICHGO:=$(shell which richgo 2>/dev/null)

clean:
	@rm -Rf $(OUTPUT_DIR)
	@mkdir $(OUTPUT_DIR)
	@mkdir $(OUTPUT_DIR_APP)
	@mkdir $(OUTPUT_DIR_MANIFESTS)
	@mkdir $(OUTPUT_DIR_TESTS)

restore:
	@cd src && go mod download -x

build:
	@cd src && GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_BUILD=0 go build \
		-v \
		-ldflags='-extldflags=-static -w -s' \
		-tags netgo,osusergo \
		-o $(OUTPUT_DIR_APP)/confluent-gateway \
		./cmd/main

# NOTE: if CGO_BUILD=0 becomes a problem down the line
# go build -ldflags='-extldflags=-static -w -s' -tags netgo,osusergo

test: tests
tests: test-setup test-scripts test-teardown
test-scripts:
	-@cd src && go test -v -cover github.com/dfds/confluent-gateway/functional_tests
	-@cd src && go test -v -cover ./...
test-setup:
	docker compose -f azure_pipelines/docker-compose.functional_tests.yml up -d --build
	@sleep 5
test-teardown:
	docker compose -f azure_pipelines/docker-compose.functional_tests.yml down

container:
	@docker build -t $(APP_IMAGE_NAME) .
	@docker build -t ${DB_IMAGE_NAME} ./db

manifests:
	@cp -r ./k8s/. $(OUTPUT_DIR_MANIFESTS)
	@find "$(OUTPUT_DIR_MANIFESTS)" -type f -name '*.yml' | xargs sed -i 's:{{BUILD_NUMBER}}:${BUILD_NUMBER}:g'

deliver:
	@sh ./tools/push-container.sh "${APP_IMAGE_NAME}" "${BUILD_NUMBER}"
	@sh ./tools/push-container.sh "${DB_IMAGE_NAME}" "${BUILD_NUMBER}"

ci: clean restore build tests container manifests
cd: ci deliver



clean-restore-build: clean restore build
docker-build-push: container deliver



run:
	@cd src && go run ./cmd/main
