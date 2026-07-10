# Build the guacamole binary
FROM docker.io/library/golang:1.26.5@sha256:63f132d58c1f589f0dcda584933a9bb44bfda1150f1506377f5a902f34d86033 as builder
ARG TARGETOS
ARG TARGETARCH
ARG PACKAGE=github.com/padok-team/guacamole
ARG VERSION
ARG COMMIT_HASH
ARG BUILD_TIMESTAMP

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY checks/ checks/
COPY cmd/ cmd/
COPY data/ data/
COPY helpers/ helpers/
COPY internal/ internal/
COPY main.go main.go

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a \
  -ldflags="\
  -X ${PACKAGE}/internal/version.Version=${VERSION} \
  -X ${PACKAGE}/internal/version.CommitHash=${COMMIT_HASH} \
  -X ${PACKAGE}/internal/version.BuildTimestamp=${BUILD_TIMESTAMP}" \
  -o bin/guacamole main.go

FROM docker.io/library/alpine:3.23.5@sha256:fd791d74b68913cbb027c6546007b3f0d3bc45125f797758156952bc2d6daf40

WORKDIR /home/guacamole

ARG TARGETARCH
ARG TERRAGRUNT_VERSION
ARG TERRAFORM_VERSION

# Install ca-certificates for HTTPS, plus terragrunt and terraform if versions are provided
RUN apk add --no-cache ca-certificates git unzip && \
    if [ -n "${TERRAGRUNT_VERSION:-}" ]; then \
      wget -qO /usr/local/bin/terragrunt \
        "https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_${TARGETARCH:-amd64}" && \
      chmod +x /usr/local/bin/terragrunt; \
    fi && \
    if [ -n "${TERRAFORM_VERSION:-}" ]; then \
      wget -qO /tmp/terraform.zip \
        "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH:-amd64}.zip" && \
      unzip -q /tmp/terraform.zip -d /usr/local/bin/ && \
      rm /tmp/terraform.zip; \
    fi

ENV UID=65532
ENV GID=65532
ENV USER=guacamole
ENV GROUP=guacamole

# Create a non-root user to run the app
RUN addgroup \
  -g $GID \
  $GROUP && \
  adduser \
  --disabled-password \
  --no-create-home \
  --home $(pwd) \
  --uid $UID \
  --ingroup $GROUP \
  $USER

# Copy the binary to the production image from the builder stage
COPY --from=builder /workspace/bin/guacamole /usr/local/bin/guacamole

RUN chmod +x /usr/local/bin/guacamole

# Use an unprivileged user
USER 65532:65532

# Run guacamole on container startup
ENTRYPOINT ["guacamole"]
