FROM node:19.6 as base
ARG BUILDARCH
SHELL ["/bin/bash", "-c"]

RUN apt-get update -qq && apt-get install -y curl unzip
RUN curl -fsSL https://bun.sh/install | bash
ENV BUN_INSTALL="/root/.bun"
ENV PATH="$BUN_INSTALL/bin:$PATH"

# Install protoc
ARG PROTOC_VERSION=21.12
RUN export ARCH=$(if [ "$BUILDARCH" == "arm64" ]; then echo "aarch_64"; else echo "x86_64"; fi); \
    wget https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-${ARCH}.zip \
    && unzip protoc-${PROTOC_VERSION}-linux-${ARCH}.zip -d /root/protoc \
    && rm protoc-${PROTOC_VERSION}-linux-${ARCH}.zip
ENV PATH="$PATH:/root/protoc/bin"

WORKDIR /opt

COPY package.json ./package.json
RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install
# Need to compile parcel if we are on arm64
RUN if [ "$BUILDARCH" == "arm64" ]; then cd node_modules/@parcel/watcher && yarn run rebuild; fi

COPY . .

ENV ESM_ROOT="./node_modules/monaco-editor/esm/vs"

FROM base as twirpscript
RUN npx twirpscript

FROM base as parcel
RUN --mount=type=cache,target=/opt/.parcel-cache,id=parcel-cache \
    npx parcel build --public-url=/assets ./_assets.go.html \
    && npx parcel build \
        "$ESM_ROOT/language/typescript/ts.worker.js" \
        "$ESM_ROOT/language/json/json.worker.js" \
        "$ESM_ROOT/language/css/css.worker.js" \
        "$ESM_ROOT/language/html/html.worker.js" \
        "$ESM_ROOT/editor/editor.worker.js" --no-source-maps

FROM node:19.6
COPY --from=parcel /opt/dist /opt/dist
COPY --from=twirpscript /opt/steady.pb.ts /opt/steady.pb.ts
