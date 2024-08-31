### BASE - USED IN ALL STEPS
################################################################################
FROM golang:1.22.2-bullseye AS base
WORKDIR /src
ENV CGO_ENABLED=1
ENV GOOS=linux

# Install ffmpeg
RUN apt-get update && apt-get install -y ffmpeg && rm -rf /var/lib/apt/lists/*

### BUILD DEPENDENCIES
################################################################################
FROM base AS build-dependencies
COPY go.mod go.sum ./
RUN go mod download

### BUILDER
################################################################################
FROM build-dependencies AS builder
COPY . .
RUN go version
RUN go env
RUN go list -m all
RUN go build -v -o /app/media-privacy-server ./cmd/webserver/main.go

### PRODUCTION SERVER
################################################################################
FROM debian:bullseye-slim AS production
RUN apt-get update && apt-get install -y ca-certificates ffmpeg && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/media-privacy-server .
COPY --from=builder /src/templates ./templates

# Expose the port the app runs on
EXPOSE 8080

# Use a non-root user
RUN useradd -m appuser
RUN chown -R appuser:appuser /app
USER appuser

ENTRYPOINT ["./media-privacy-server"]