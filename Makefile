don: *.go
	go build -o don

.PHONY: clean

clean:
	rm -f don
