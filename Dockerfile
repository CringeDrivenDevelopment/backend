# Base image with go compiler
FROM golang:1.25.4 AS build

# Go to working dir
WORKDIR /src

# Install build tools
RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt/lists \
    apt-get update -y && \
    apt-get install -y --no-install-recommends \
        git \
        ca-certificates \
        && \
    rm -rf /var/lib/apt/lists/*

# Copy module files
COPY go.mod go.sum ./

# Install deps
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download -x

# Copy source code
COPY . .

# Build app
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 go build -o /bin/server ./cmd/main.go

# Final stage
FROM debian:trixie-slim AS final

# Set timezone and locale
ENV TZ=UTC \
    LANG=C.UTF-8

# Install runtime deps
RUN --mount=type=cache,sharing=locked,target=/var/cache/apt \
    --mount=type=cache,sharing=locked,target=/var/lib/apt/lists \
    rm -rf /var/lib/apt/lists/* && \
    apt-get update -y && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        tzdata \
        curl \
        && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user for security
ARG UID=10001
RUN useradd -u ${UID} \
    -r \
    -g root \
    -d /nonexistent \
    -s /sbin/nologin \
    -c "Application user" \
    appuser
USER appuser

# Copy binary
COPY --from=build --chown=appuser:appuser /bin/server /bin/

# Set run permissions
RUN chmod +x /bin/server

# Expose application port
EXPOSE 8080

# Entrypoint
ENTRYPOINT ["/bin/server"]