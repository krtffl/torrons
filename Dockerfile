FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

ARG TARGETARCH

# Install git for version detection in Makefile
RUN apk add --no-cache git make

WORKDIR /app

COPY . /app/

RUN if [ "$TARGETARCH" = "arm64" ]; then make dist-arm64 ; else make dist ; fi

FROM alpine:3.21

# Install ca-certificates for HTTPS connections
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/out/* /app/

WORKDIR /app

ENTRYPOINT [ "./server" ]
