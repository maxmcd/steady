

# We pin to a fixed release, should update periodically as desired.
# https://github.com/NixOS/nixpkgs/commit/c95bf18beba4290af25c60cbaaceea1110d0f727
with (import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/c95bf18beba4290af25c60cbaaceea1110d0f727.tar.gz") {});
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
