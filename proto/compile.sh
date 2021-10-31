#!/usr/bin/env sh

## Protobuf
# https://developers.google.com/protocol-buffers/docs/gotutorial
# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#
## GRPC
# https://grpc.io/docs/languages/go/quickstart/
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

rm -f ./filecrypt.pb.go
protoc --go_out=./ --go-grpc_out=. ./filecrypt.proto
