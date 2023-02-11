package web

//go:generate bash -c "bun install || bun install"

//go:generate ./generate_ts_proto.sh

//go:generate bun x parcel build ./templates/_assets.go.html
