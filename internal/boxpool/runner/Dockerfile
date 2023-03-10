FROM debian:11 as bun

RUN useradd -u 1500 -ms /bin/bash steady
RUN mkdir -p /home/steady
RUN apt-get update -qq && apt-get install -y curl unzip

USER steady
RUN curl -fsSL https://bun.sh/install | bash
ENV BUN_INSTALL="/home/steady/.bun"
ENV PATH="$BUN_INSTALL/bin:$PATH"
# Install supporting files needed for bun repl.
RUN bun add -g bun-repl
# Patch the location of the history log
COPY ./internal/boxpool/runner/replmanager.ts /home/steady/.bun/install/global/node_modules/bun-repl/src/replmanager.ts


FROM golang:1.20-bullseye as go

WORKDIR /steady

COPY ./go.mod ./go.sum /steady/
COPY ./internal/boxpool/*.go /steady/internal/boxpool/
COPY ./internal/boxpool/runner/*.go /steady/internal/boxpool/runner/
COPY ./internal/slogx/*.go /steady/internal/slogx/
COPY ./internal/execx/*.go /steady/internal/execx/
COPY ./internal/steadyutil/*.go /steady/internal/steadyutil/

WORKDIR /steady/internal/boxpool/runner
RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    --mount=type=cache,id=gomcache,target=/root/.cache/go-build \
    go build .

FROM debian:11

RUN apt-get update -qq && apt-get install -y strace

ENV BUN_INSTALL="/home/steady/.bun"
ENV PATH="$BUN_INSTALL/bin:$PATH"
ENV HOME="/home/steady"

COPY --from=go /steady/internal/boxpool/runner/runner /bin/runner
# https://github.com/GoogleContainerTools/distroless/issues/277#issuecomment-456941869
COPY --from=bun /etc/passwd /etc/passwd
COPY --from=bun /home/steady /home/steady
COPY --from=bun /home/steady/.bun /home/steady/.bun
COPY --from=bun /tmp /tmp
COPY ./internal/boxpool/runner/wrapper.ts /home/steady/wrapper.ts

# VOLUME /tmp
# USER steady

ENTRYPOINT []
CMD ["/bin/runner"]
