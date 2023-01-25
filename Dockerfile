# Build the manager binary
FROM  --platform=$BUILDPLATFORM golang:1.19 as builder
ARG TARGETOS
ARG TARGETARCH


WORKDIR /workspace


# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer

COPY sk-common/go.mod sk-common/go.mod
COPY sk-common/go.sum sk-common/go.sum
RUN cd sk-common && go mod download

COPY sk-crd/go.mod sk-crd/go.mod
COPY sk-crd/go.sum sk-crd/go.sum
RUN cd sk-crd && go mod download

COPY sk-ldap/go.mod sk-ldap/go.mod
COPY sk-ldap/go.sum sk-ldap/go.sum
RUN cd sk-ldap && go mod download

COPY sk-static/go.mod sk-static/go.mod
COPY sk-static/go.sum sk-static/go.sum
RUN cd sk-static && go mod download

# Copy and build go programs

COPY sk-common/pkg/ sk-common/pkg/
COPY sk-common/proto/ sk-common/proto/

COPY sk-crd/internal/ sk-crd/internal/
COPY sk-crd/k8sapis/ sk-crd/k8sapis/
COPY sk-crd/main.go sk-crd/main.go
RUN cd sk-crd && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o sk-crd main.go

COPY sk-ldap/internal/ sk-ldap/internal/
COPY sk-ldap/main.go sk-ldap/main.go
RUN cd sk-ldap && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o sk-ldap main.go

COPY sk-static/internal/ sk-static/internal/
COPY sk-static/main.go sk-static/main.go
RUN cd sk-static && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o sk-static main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/sk-crd/sk-crd .
COPY --from=builder /workspace/sk-ldap/sk-ldap .
COPY --from=builder /workspace/sk-static/sk-static .
USER 65532:65532

#ENTRYPOINT [""]
