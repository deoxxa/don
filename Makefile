don: *.go
	go build -o don
	rice append --exec don

.PHONY: clean

clean:
	rm -f don
