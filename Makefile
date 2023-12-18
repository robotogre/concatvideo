build:
	go build

run4k: build
	./concatvideo -res 4K -s3 true