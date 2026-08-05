# ── Stage 1: Build the Go binary ──────────────────────────
FROM golang:1.25-alpine AS builder

# Install git for module downloads (Alpine is minimal)
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependencies — copy go.mod / go.sum first so Docker
# can cache this layer and avoid re-downloading on every build.
COPY go.mod go.sum* ./
RUN go mod download

# Copy the rest of the source and compile a statically-linked binary.
# CGO_ENABLED=0 produces a binary with no libc dependency, which is
# required for the scratch/distroless runtime image.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# ── Stage 2: Minimal runtime image ───────────────────────────
# distroless-static contains ca-certificates and nothing else —
# no shell, no package manager. Small attack surface.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /server /server

# Copy seed data so the binary can find it at /app/data/seed_data.json
COPY --from=builder /app/data /app/data

# Run as non-root user (provided by distroless)
EXPOSE 8080
ENTRYPOINT ["/server"]
