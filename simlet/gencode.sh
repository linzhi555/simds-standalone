go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
./protoc_tools/linux/protoc --go_out=.  --go-grpc_out=require_unimplemented_servers=false:. ./simlet.proto
