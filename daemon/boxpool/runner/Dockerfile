FROM golang

WORKDIR /opt

RUN useradd -ms /bin/bash steady
RUN mkdir -p /home/newuser

COPY . .

RUN go mod init opt && go mod tidy
RUN go mod download
RUN go build .

FROM gcr.io/distroless/base-debian11:debug

COPY --from=0 /opt/opt /bin/runner
# https://github.com/GoogleContainerTools/distroless/issues/277#issuecomment-456941869
COPY --from=0 /etc/passwd /etc/passwd
COPY --from=0 /home/newuser /home/newuser

# USER steady

ENTRYPOINT []
CMD ["/bin/runner"]
