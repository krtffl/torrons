FROM --platform=$BUILDPLATFORM golang:1.25-alpine@sha256:1ae0735f00daffa3aaf1363a5184c0d2dc55c78e3db4ec70241cdac97bf84b59 AS builder

ARG TARGETARCH

# Install git for version detection in Makefile
RUN apk add --no-cache git make

WORKDIR /app

COPY . /app/

RUN if [ "$TARGETARCH" = "arm64" ]; then make dist-arm64 ; else make dist ; fi

FROM alpine:3.21@sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d

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
