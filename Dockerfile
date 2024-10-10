FROM --platform=$BUILDPLATFORM golang:1.23.2-alpine AS build

ARG VERSION
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X main.version=${VERSION}" -o dist/kiosk .


FROM  alpine:latest

ENV TZ=Europe/London

ENV TERM=xterm-256color
ENV DEBUG_COLORS=true
ENV COLORTERM=truecolor

RUN apk update && apk add --no-cache tzdata ca-certificates && update-ca-certificates

WORKDIR /

COPY --from=build /app/dist/kiosk .

ENTRYPOINT ["/kiosk"]
