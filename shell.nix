with (import <nixpkgs> {});
mkShell {
    # nativeBuildInputs is usually what you want -- tools you need to run
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
    ];
}
