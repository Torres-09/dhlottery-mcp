FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /dhlottery-mcp ./cmd/dhlottery-mcp

FROM scratch
COPY --from=builder /dhlottery-mcp /dhlottery-mcp
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/dhlottery-mcp"]
