# Build the manager binary
FROM golang:1.19.0 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod koomgr/go.mod
COPY go.sum koomgr/go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN cd koomgr; go mod download

# Copy the go source
COPY main.go koomgr/main.go
COPY apis koomgr/apis/
COPY internal koomgr/internal/

WORKDIR /workspace/koomgr

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o /workspace/koomgr/manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/koomgr/manager .
# Must be numerical, for PSP (Checked in /etc/password of the image)
USER 65532:65532

ENTRYPOINT ["/manager"]
