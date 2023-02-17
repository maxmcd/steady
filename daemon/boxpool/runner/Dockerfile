FROM debian:11 as bun

RUN apt-get update -qq && apt-get install -y curl unzip
RUN curl -fsSL https://bun.sh/install | bash


FROM debian:11 as user

RUN useradd -ms /bin/bash steady
RUN mkdir -p /home/newuser

FROM golang as golang

WORKDIR /steady
RUN mkdir -p /steady/daemon/boxpool/runner

COPY ./go.mod ./go.sum /steady/
COPY ./daemon/boxpool/*.go /steady/daemon/boxpool
COPY ./daemon/boxpool/runner/*.go /steady/daemon/boxpool/runner

WORKDIR /steady/daemon/boxpool/runner
RUN go build .

FROM gcr.io/distroless/base-debian11:debug

ENV BUN_INSTALL="/root/.bun"
ENV PATH="$BUN_INSTALL/bin:$PATH"

COPY --from=golang /steady/daemon/boxpool/runner/runner /bin/runner
# https://github.com/GoogleContainerTools/distroless/issues/277#issuecomment-456941869
COPY --from=user /etc/passwd /etc/passwd
COPY --from=user /home/newuser /home/newuser
COPY --from=bun /root/.bun /root/.bun

# USER steady

ENTRYPOINT []
CMD ["/bin/runner"]
