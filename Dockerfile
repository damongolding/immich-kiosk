FROM --platform=$BUILDPLATFORM golang:1.22.5-alpine AS build

ARG VERSION
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY . .

RUN go mod download
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o dist/frame .


FROM  alpine:latest

WORKDIR /

COPY --from=build /app/dist/frame .

EXPOSE 3000

ENTRYPOINT ["/frame"]
