FROM golang as builder

ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0

RUN go install github.com/protoc-extensions/protoc-gen-gapi-lint/cmd/protoc-gen-gapi-lint@latest
# Note, the Docker images must be built for amd64. If the host machine architecture is not amd64
# you need to cross-compile the binary and move it into /go/bin.
RUN bash -c 'find /go/bin/${GOOS}_${GOARCH}/ -mindepth 1 -maxdepth 1 -exec mv {} /go/bin \;'

FROM scratch

LABEL "build.buf.plugins.runtime_library_versions.0.name"="github.com/protoc-extensions/protoc-gen-gapi-lint"
LABEL "build.buf.plugins.runtime_library_versions.0.version"="v0.0.0-20230408142104-eac27b75b5a6"
LABEL "build.buf.plugins.runtime_library_versions.1.name"="google.golang.org/protobuf"
LABEL "build.buf.plugins.runtime_library_versions.1.version"="v1.29.0"

COPY --from=builder /go/bin /

ENTRYPOINT ["/protoc-gen-gapi-lint"]
