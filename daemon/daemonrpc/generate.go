package daemonrpc

//go:generate protoc --proto_path=. --go_out=. --twirp_out=. --twirp_opt=paths=source_relative --go_opt=paths=source_relative ./daemon.proto
