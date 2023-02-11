package web

//go:generate bun install

//go:generate ./generate_ts_proto.sh

//go:generate bunx parcel build ./templates/_assets.go.html
