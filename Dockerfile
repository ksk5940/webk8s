# ---------- build stage ----------
FROM golang:1.22 AS builder

WORKDIR /src

# Copy go mod first for caching
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /src/backend
RUN go mod download

# Copy full source
WORKDIR /src
COPY backend ./backend
COPY ui ./ui

# Build binary
WORKDIR /src/backend
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/webk8s ./cmd/server

# ---------- runtime stage ----------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=builder /out/webk8s /webk8s
COPY --from=builder /src/ui /ui

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/webk8s"]
