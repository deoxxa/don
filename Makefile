don: *.go $(shell find migrations public templates -type f) build/entry-server-bundle.js build/entry-client-bundle.js
	@echo "--> Building don with host toolchain"
	@go build -v -ldflags=-s -o don
	@rice append --exec don

build/entry-server-bundle.js: $(shell find client/src -type f) client/webpack.* client/yarn.lock client/package.json
	@echo "--> Building server JavaScript bundle"
	@docker run --rm -it -v `pwd`:/app deoxxa/node-yarn:7.4 bash -c 'cd /app/client && yarn run build-server'
	@touch build/entry-server-bundle.js

build/entry-client-bundle.js: $(shell find client/src -type f) client/webpack.* client/yarn.lock client/package.json
	@echo "--> Building client JavaScript bundle"
	@docker run --rm -it -v `pwd`:/app deoxxa/node-yarn:7.4 bash -c 'cd /app/client && yarn run build-client'
	@touch build/entry-client-bundle.js

cross.stamp: *.go $(shell find migrations public templates -type f) build/entry-server-bundle.js build/entry-client-bundle.js
	@echo "--> Building cross-platform binaries"
	@xgo -targets 'darwin/amd64,linux/amd64,linux/arm,windows/amd64' .
	@rice append --exec don-darwin-10.6-amd64
	@rice append --exec don-linux-amd64
	@rice append --exec don-linux-arm-5
	@rice append --exec don-windows-4.0-amd64.exe
	@touch cross.stamp

.PHONY: cross release clean

BINTRAY_USER ?= deoxxa
BINTRAY_REPO ?= don
BINTRAY_PACKAGE ?= portable
BINTRAY_VERSION ?= dev

cross: cross.stamp

clean:
	@echo "--> Removing build artifacts"
	@rm -rvf cross.stamp don don-darwin-10.6-amd64 don-linux-amd64 don-linux-arm-5 don-windows-4.0-amd64.exe build/*

release: cross.stamp
ifeq ($(BINTRAY_VERSION) , dev)
	@echo "--> Removing dev Version"
	@curl --silent --request DELETE --user "${BINTRAY_AUTH}" "https://api.bintray.com/packages/${BINTRAY_USER}/${BINTRAY_REPO}/${BINTRAY_PACKAGE}/versions/${BINTRAY_VERSION}"
	@echo
endif
	@echo "--> Uploading version: ${BINTRAY_VERSION}"
	@echo "... darwin-amd64"
	@curl --output /dev/null --progress-bar --upload-file don-darwin-10.6-amd64     --user "${BINTRAY_AUTH}" "https://api.bintray.com/content/${BINTRAY_USER}/${BINTRAY_REPO}/${BINTRAY_PACKAGE}/${BINTRAY_VERSION}/${BINTRAY_REPO}_${BINTRAY_VERSION}_darwin-amd64"
	@echo "... linux-amd64"
	@curl --output /dev/null --progress-bar --upload-file don-linux-amd64           --user "${BINTRAY_AUTH}" "https://api.bintray.com/content/${BINTRAY_USER}/${BINTRAY_REPO}/${BINTRAY_PACKAGE}/${BINTRAY_VERSION}/${BINTRAY_REPO}_${BINTRAY_VERSION}_linux-amd64"
	@echo "... linux-arm"
	@curl --output /dev/null --progress-bar --upload-file don-linux-arm-5           --user "${BINTRAY_AUTH}" "https://api.bintray.com/content/${BINTRAY_USER}/${BINTRAY_REPO}/${BINTRAY_PACKAGE}/${BINTRAY_VERSION}/${BINTRAY_REPO}_${BINTRAY_VERSION}_linux-arm"
	@echo "... windows-amd64"
	@curl --output /dev/null --progress-bar --upload-file don-windows-4.0-amd64.exe --user "${BINTRAY_AUTH}" "https://api.bintray.com/content/${BINTRAY_USER}/${BINTRAY_REPO}/${BINTRAY_PACKAGE}/${BINTRAY_VERSION}/${BINTRAY_REPO}_${BINTRAY_VERSION}_windows-amd64.exe"
	@echo "--> Publishing version ${BINTRAY_VERSION}"
	@curl --silent --request POST --user "${BINTRAY_AUTH}" "https://api.bintray.com/content/${BINTRAY_USER}/${BINTRAY_REPO}/${BINTRAY_PACKAGE}/${BINTRAY_VERSION}/publish"
	@echo
	@echo "--> Waiting a little while for bintray to completely publish ${BINTRAY_VERSION}"
	@echo "... 30s"
	@sleep 10
	@echo "... 20s"
	@sleep 10
	@echo "... 10s"
	@sleep 10
	@echo "--> Making files visible"
	@echo "... darwin-amd64"
	@curl --output /dev/null --silent --user "${BINTRAY_AUTH}" --request PUT --header 'content-type: application/json' --data-binary '{"list_in_downloads": true}' "https://api.bintray.com/file_metadata/${BINTRAY_USER}/${BINTRAY_REPO}/${BINTRAY_REPO}_${BINTRAY_VERSION}_darwin-amd64"
	@echo "... linux-amd64"
	@curl --output /dev/null --silent --user "${BINTRAY_AUTH}" --request PUT --header 'content-type: application/json' --data-binary '{"list_in_downloads": true}' "https://api.bintray.com/file_metadata/${BINTRAY_USER}/${BINTRAY_REPO}/${BINTRAY_REPO}_${BINTRAY_VERSION}_linux-amd64"
	@echo "... linux-arm"
	@curl --output /dev/null --silent --user "${BINTRAY_AUTH}" --request PUT --header 'content-type: application/json' --data-binary '{"list_in_downloads": true}' "https://api.bintray.com/file_metadata/${BINTRAY_USER}/${BINTRAY_REPO}/${BINTRAY_REPO}_${BINTRAY_VERSION}_linux-arm"
	@echo "... windows-amd64"
	@curl --output /dev/null --silent --user "${BINTRAY_AUTH}" --request PUT --header 'content-type: application/json' --data-binary '{"list_in_downloads": true}' "https://api.bintray.com/file_metadata/${BINTRAY_USER}/${BINTRAY_REPO}/${BINTRAY_REPO}_${BINTRAY_VERSION}_windows-amd64.exe"
	@echo "--> Release is at https://bintray.com/${BINTRAY_USER}/${BINTRAY_REPO}/${BINTRAY_PACKAGE}/${BINTRAY_VERSION}"
