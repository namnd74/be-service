# Stage 1: Build a static binary for the requested target architecture.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS builder

ARG APP_VERSION=v1.1.0
ARG GIT_COMMIT=dev-local

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
ARG TARGETARCH
# Opt-in seminar fault: build fails before an image can be published.
ARG DEMO_BUILD_FAIL=false
RUN if [ "$DEMO_BUILD_FAIL" = true ]; then echo "SEMINAR: intentional prod build failure" >&2; exit 1; fi
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s -X main.buildVersion=${APP_VERSION} -X main.buildCommit=${GIT_COMMIT}" \
    -o server .

# Stage 2: No OS packages in the runtime image.
FROM scratch

ARG APP_VERSION=v1.1.0
ARG GIT_COMMIT=dev-local
ARG SOURCE_URL=local://be-service
LABEL org.opencontainers.image.version=$APP_VERSION \
      org.opencontainers.image.revision=$GIT_COMMIT \
      org.opencontainers.image.source=$SOURCE_URL

USER 65532:65532

WORKDIR /
COPY --from=builder /app/server /server

ENV PORT=8080 \
    APP_ENV=dev \
    DEMO_MODE=false \
    DEMO_FAULT=false

EXPOSE 8080

ENTRYPOINT ["/server"]
