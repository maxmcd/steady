FROM node:19.6

RUN apt-get update -qq && apt-get install -y curl unzip
RUN curl -fsSL https://bun.sh/install | bash
ENV BUN_INSTALL="/root/.bun"
ENV PATH="$BUN_INSTALL/bin:$PATH"

ARG PROTOC_VERSION=21.12
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip \
    && unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /root/protoc \
    && rm protoc-${PROTOC_VERSION}-linux-x86_64.zip
ENV PATH="$PATH:/root/protoc/bin"

WORKDIR /opt
COPY package.json ./package.json
RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install
COPY . .
RUN  --mount=type=cache,target=/opt/.parcel-cache \
    npx parcel build --public-url=/assets ./templates/_assets.go.html

RUN npx twirpscript
