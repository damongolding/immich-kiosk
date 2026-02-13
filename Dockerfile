# Frontend Base Image
FROM oven/bun:1 AS frontend-base

COPY . /app
WORKDIR /app/frontend

FROM frontend-base AS frontend-build
RUN bun install --frozen-lockfile
RUN bun run css && bun run js && bun run url-builder


# Go Builder
FROM --platform=$BUILDPLATFORM golang:1.26.0-bookworm AS build

ARG VERSION=demo
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY . .
COPY --from=frontend-build /app/frontend/public/assets/css /app/frontend/public/assets/css
COPY --from=frontend-build /app/frontend/public/assets/js/ /app/frontend/public/assets/js/

RUN go mod download
RUN go tool templ generate

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -installsuffix cgo -ldflags "-X main.version=${VERSION}" -o dist/kiosk .

# Release
FROM alpine:3.22.2

ENV TZ=Europe/London

ENV TERM=xterm-256color
ENV DEBUG_COLORS=true
ENV COLORTERM=truecolor

RUN apk add --no-cache tzdata ca-certificates curl && update-ca-certificates

WORKDIR /

COPY --from=build /app/demo.config.yaml .
COPY --from=build /app/dist/kiosk .

ENTRYPOINT ["/kiosk"]
