FROM golang:1.17-bullseye as builder
WORKDIR /go/src
COPY . /go/src
RUN set -e \
    && go mod download \
    && go build -ldflags "-X main.version=$REVISION" -o openpoker-backend cmd/main.go

FROM debian:bullseye
WORKDIR /opt/openpoker
COPY --from=builder /go/src/openpoker-backend .
RUN set -e \
    && apt update -yqq \
    && apt install -yqq ca-certificates \
    && apt clean autoclean \
    && apt autoremove -yqq \
    && rm -rf /var/lib/apt/lists/*
EXPOSE 8080
CMD ["/opt/openpoker/openpoker-backend"]