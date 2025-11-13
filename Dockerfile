# Frontend Base Image
FROM node:22-slim AS frontend-base

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
COPY . /app
WORKDIR /app/frontend

# Frontend Dependencies
FROM frontend-base AS frontend-prod-deps
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --prod --frozen-lockfile

# Frontend Build
FROM frontend-base AS frontend-build
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
RUN pnpm css && pnpm js && pnpm url-builder

# Go Builder
FROM --platform=$BUILDPLATFORM golang:1.25.4-alpine AS build

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
FROM alpine:latest

ENV TZ=Europe/London

ENV TERM=xterm-256color
ENV DEBUG_COLORS=true
ENV COLORTERM=truecolor

RUN apk add --no-cache tzdata ca-certificates curl && update-ca-certificates

WORKDIR /

COPY --from=build /app/demo.config.yaml .
COPY --from=build /app/dist/kiosk .

ENTRYPOINT ["/kiosk"]
