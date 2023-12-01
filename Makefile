#!/usr/bin/make -f

syntryve:
	go build -mod=readonly -o ./build/syntryve ./cmd/syntryve/main.go