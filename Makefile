APP_IMAGE_NAME=selfservice/confluent-gateway
BUILD_NUMBER=n/a
OUTPUT_DIR=${PWD}/.output
OUTPUT_DIR_APP=${OUTPUT_DIR}/app
OUTPUT_DIR_MANIFESTS=${OUTPUT_DIR}/manifests
GOOS=linux
GOARCH=amd64

clean:
	@rm -Rf $(OUTPUT_DIR)
	@mkdir $(OUTPUT_DIR)
	@mkdir $(OUTPUT_DIR_APP)
	@mkdir $(OUTPUT_DIR_MANIFESTS)

restore:
	@cd src && go mod download -x

build:
	@cd src && GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-v \
		-ldflags '-w -s' \
		-o $(OUTPUT_DIR_APP)/confluent-gateway

container:
	@docker build -t $(APP_IMAGE_NAME) .

manifests:
	@cp -r ./k8s/. $(OUTPUT_DIR_MANIFESTS)
	@find "$(OUTPUT_DIR_MANIFESTS)" -type f -name '*.yml' | xargs sed -i 's:{{BUILD_NUMBER}}:${BUILD_NUMBER}:g'

deliver:
	@sh ./tools/push-container.sh "${APP_IMAGE_NAME}" "${BUILD_NUMBER}"

ci: clean restore build container
cd: ci deliver