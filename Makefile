don: *.go migrations.rice-box.go public.rice-box.go templates.rice-box.go
	go build -ldflags=-s -o don

don_linux-amd64: *.go public.rice-box.go
	docker run --rm -it -v $(shell pwd):/go/src/fknsrs.biz/p/don golang:1.8 bash -c 'cd /go/src/fknsrs.biz/p/don && go build -o don_linux-amd64'

build/entry-server-bundle.js: $(shell find client/src -type f) client/webpack.* client/yarn.lock client/package.json
	cd client && yarn run build-server

build/entry-client-bundle.js: $(shell find client/src -type f) client/webpack.* client/yarn.lock client/package.json
	cd client && yarn run build-client

.PHONY: clean

clean:
	rm -rvf don don_linux-amd64 build/*
