FROM golang:1.17.3-alpine as builder

RUN apk update && \
    apk add --no-cache git ca-certificates && \
    update-ca-certificates

WORKDIR /go/src/report

COPY . .
RUN go mod download

ARG CMD_BIN
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -mod=readonly -a -o app $CMD_BIN

FROM scratch
WORKDIR /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/report .
ENTRYPOINT ["/app"]
