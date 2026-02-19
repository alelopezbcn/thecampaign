# Build stage
FROM golang:1.25-alpine AS builder

ARG VERSION=dev

WORKDIR /build/backend

# Copy backend source with vendor
COPY backend/ .

# Build static binary (no CGO needed)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor \
    -ldflags "-X 'github.com/alelopezbcn/thecampaign/internal/server.Version=${VERSION}'" \
    -o /build/server ./cmd/server

# Runtime stage
FROM alpine:3.21

WORKDIR /app

# Copy binary and frontend
COPY --from=builder /build/server .
COPY frontend/ frontend/

EXPOSE 8080

ENV PORT=8080

CMD ["./server"]
