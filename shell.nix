

# We pin to a fixed release, should update periodically as desired.
# https://github.com/NixOS/nixpkgs/commit/988cc958c57ce4350ec248d2d53087777f9e1949
with (import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/988cc958c57ce4350ec248d2d53087777f9e1949.tar.gz") {});
mkShell {
    nativeBuildInputs = [
        temporalite
        bun
        minio
        go
        python310Packages.codecov
        sqlc
        protobuf
        protoc-gen-go
        protoc-gen-twirp
        golangci-lint
    ];
}
