// Server
//go:generate go run github.com/webrpc/webrpc/cmd/webrpc-gen -schema=rpc.ridl -target=golang -pkg=proto -server -client -out=./rpc.gen.go

// Clients
//
//go:generate go run github.com/webrpc/webrpc/cmd/webrpc-gen -schema=rpc.ridl -target=golang -pkg=bundler -client -out=./clients/proto.gen.go
package proto
