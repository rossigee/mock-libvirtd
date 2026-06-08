FROM golang:1.26.4 as builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b /usr/local/bin v2.12.2
RUN golangci-lint run ./...

RUN go test -v -race -coverprofile=coverage.out ./...

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o app ./cmd/main

FROM scratch
COPY --from=builder /build/app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8080
ENTRYPOINT ["/app"]
