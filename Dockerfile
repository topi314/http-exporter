FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG COMMIT

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -ldflags="-X 'main.Version=$VERSION' -X 'main.Commit=$COMMIT'" -o prometheus-collectors github.com/topi314/prometheus-collectors

FROM alpine

RUN apk add --no-cache  \
    inkscape \
    ttf-freefont

COPY --from=build /build/prometheus-collectors /bin/prometheus-collectors

EXPOSE 80

ENTRYPOINT ["/bin/prometheus-collectors"]

CMD ["-config", "/var/lib/prometheus-collectors/config.toml"]
