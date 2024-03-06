// Server
//go:generate go run github.com/webrpc/webrpc/cmd/webrpc-gen@v0.14.2 -schema=rpc.ridl -target=golang -pkg=proto -server -client -out=./rpc.gen.go

// Clients
//
//go:generate go run github.com/webrpc/webrpc/cmd/webrpc-gen@v0.14.2 -schema=rpc.ridl -target=golang -pkg=bundler -client -out=./clients/proto.gen.go
//go:generate go run github.com/webrpc/webrpc/cmd/webrpc-gen@v0.14.2 -schema=rpc.ridl -target=typescript -client -out=./clients/proto.gen.ts

package proto
