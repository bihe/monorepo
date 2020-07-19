#!/usr/bin/env sh

## taken from go-kit
## https://github.com/go-kit/kit/blob/master/examples/addsvc/pb/compile.sh

# Install proto3 from source
#  brew install autoconf automake libtool
#  git clone https://github.com/google/protobuf
#  ./autogen.sh ; ./configure ; make ; make install
#
# Update protoc Go bindings via
#  go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
#
# See also
#  https://github.com/grpc/grpc-go/tree/master/examples

rm -f ./filecrypt.pb.go
protoc filecrypt.proto --go_out=plugins=grpc:.
