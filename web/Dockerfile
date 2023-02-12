FROM debian

RUN apt-get update -qq && apt-get install -y curl unzip nodejs
RUN curl -fsSL https://bun.sh/install | bash
ENV BUN_INSTALL="/root/.bun"
ENV PATH="$BUN_INSTALL/bin:$PATH"

WORKDIR /optdocker
COPY package.json ./package.json
RUN bun install
COPY . .
RUN node npx parcel build --public-url=/assets ./templates/_assets.go.html
