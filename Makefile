don: *.go $(shell find migrations public templates -type f) build/entry-server-bundle.js build/entry-client-bundle.js
	go build -ldflags=-s -o don
	rice append --exec don

don_linux-amd64: *.go build/entry-server-bundle.js build/entry-client-bundle.js $(shell find migrations public templates -type f)
	docker run --rm -it -v $(shell pwd):/go/src/fknsrs.biz/p/don golang:1.8 bash -c 'cd /go/src/fknsrs.biz/p/don && go build -o don_linux-amd64'
	rice append --exec don_linux-amd64

build/entry-server-bundle.js: $(shell find client/src -type f) client/webpack.* client/yarn.lock client/package.json
	cd client && yarn run build-server

build/entry-client-bundle.js: $(shell find client/src -type f) client/webpack.* client/yarn.lock client/package.json
	cd client && yarn run build-client

.PHONY: clean

clean:
	rm -rvf don don_linux-amd64 build/*
