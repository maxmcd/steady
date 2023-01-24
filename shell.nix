with (import <nixpkgs> {});
mkShell {
    # nativeBuildInputs is usually what you want -- tools you need to run
    nativeBuildInputs = [
        temporalite
        bun
        minio
        go
        oapi-codegen
    ];
}
