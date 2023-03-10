

# We pin to a fixed release, should update periodically as desired.
# https://github.com/NixOS/nixpkgs/commit/38b7104fd1db0046ceed579f5dab4e62f136589c
with (import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/38b7104fd1db0046ceed579f5dab4e62f136589c.tar.gz") {});
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
