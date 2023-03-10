FROM debian:11 as base

RUN apt-get update -qq && apt-get install -y curl unzip

RUN curl -fsSL https://bun.sh/install | bash

FROM gcr.io/distroless/base-debian11:debug

ENV BUN_INSTALL="/root/.bun"
ENV PATH="$BUN_INSTALL/bin:$PATH"

COPY --from=base /root/.bun /root/.bun

CMD ["/root/.bun/bin/bun"]

