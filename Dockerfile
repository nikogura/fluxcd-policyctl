# Stage 1: Build the UI
FROM node:18-alpine AS ui-builder

WORKDIR /app/ui
COPY pkg/ui/package*.json ./
RUN npm ci

COPY pkg/ui/ ./
RUN npm run build

# Stage 2: Build the Go binary
FROM golang:1.25-alpine AS go-builder

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Copy the built UI into the Go embed directory
COPY --from=ui-builder /app/ui/dist ./pkg/ui/dist

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o fluxcd-policyctl .

# Stage 3: Runtime
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary
COPY --from=go-builder /app/fluxcd-policyctl .

# Create non-root user
RUN addgroup -g 1001 -S policyctl && \
    adduser -u 1001 -S policyctl -G policyctl

USER policyctl

EXPOSE 9999

CMD ["./fluxcd-policyctl"]
