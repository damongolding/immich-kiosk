FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build

ARG VERSION
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY . .
COPY config.example.yaml /app/config/

RUN go mod download
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X main.version=${VERSION}" -o dist/kiosk .


FROM  alpine:latest

ENV TZ=Europe/London

RUN apk add --no-cache tzdata

WORKDIR /

COPY --from=build /app/dist/kiosk .

EXPOSE 3000

ENTRYPOINT ["/kiosk"]
