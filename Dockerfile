FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETARCH

# Install git for version detection in Makefile
RUN apk add --no-cache git make

WORKDIR /app

COPY . /app/

RUN if [ "$TARGETARCH" = "arm64" ]; then make dist-arm64 ; else make dist ; fi

FROM alpine:3.21

# Install ca-certificates for HTTPS connections
RUN apk add --no-cache ca-certificates

# Run as a non-root user rather than the container default (root)
RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder --chown=app:app /app/out/* /app/

# LOGGER_PATH (config/config.yaml) defaults to logs/torro.log, relative to
# WORKDIR - the app user needs write access to create it.
RUN mkdir -p /app/logs && chown app:app /app/logs

USER app

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider "http://localhost:${PORT:-3000}/healthcheck" || exit 1

ENTRYPOINT [ "./server" ]
