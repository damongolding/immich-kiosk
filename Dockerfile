FROM --platform=$BUILDPLATFORM golang:1.22.5-alpine AS build

ARG VERSION
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY . .
COPY config.example.yaml /app/config/

RUN go mod download
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o dist/kiosk .


FROM  alpine:latest

WORKDIR /

COPY --from=build /app/dist/kiosk .
COPY public /public

EXPOSE 3000

ENTRYPOINT ["/kiosk"]
