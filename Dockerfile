# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates tzdata
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o /out/farm-manager ./cmd/farm-manager

FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=builder /out/farm-manager /farm-manager
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/farm-manager"]
