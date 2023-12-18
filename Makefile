build:
	go build

run4k: build
	./concatvideo -res 4K -s3 true

runfast: build
	./concatvideo -res HD -s3=true -fast=true
