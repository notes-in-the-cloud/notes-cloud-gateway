FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY api-gateway/go.mod api-gateway/go.sum ./
RUN go mod download

COPY api-gateway/ ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /api-gateway ./cmd/api-gateway

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

COPY --from=builder /api-gateway /api-gateway

EXPOSE 8090

ENTRYPOINT ["/api-gateway"]
