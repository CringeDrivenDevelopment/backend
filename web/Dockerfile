ARG GO_VERSION=1.24
FROM golang:${GO_VERSION} AS build
WORKDIR /src

COPY . .
RUN go mod download -x && go mod verify

# RUN apk update && apk add ca-certificates git gcc g++ libc-dev binutils

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    # CGO_ENABLED=0 \
    go build -o /bin/server ./cmd/main.go

FROM alpine:latest AS final

RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
        ca-certificates \
        tzdata \
        libcap-setcap \
        && \
        update-ca-certificates

COPY --from=build /bin/server /bin/
RUN setcap cap_net_raw+ep /bin/server

ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

EXPOSE 8080

ENTRYPOINT [ "/bin/server" ]