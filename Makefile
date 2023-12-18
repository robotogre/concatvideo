build:
	go build

run4k: build
	./concatvideo -res 4K -s3 true

runfast: build
	./concatvideo -res HD -s3=true -fast=true

runshort: build
	./concatvideo -res HD -s3=true -short=true

runshortfast: build
	./concatvideo -res HD -s3=true -short=true -fast=true

run-pro-all: build
	rm -rf theaterdemosHD.mp4
	rm -rf theaterdemos4K.mp4
	./concatvideo -res HD,4K -s3=true
