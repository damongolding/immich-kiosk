# Immich Kiosk

**This project is not affiliated with [immich][immich-github-url]**

## Table of Contents
- [What is Immich Kiosk?](#what-is-immich-kiosk)
- [Installation](#installation)
- [Configuration](#configuration)
- [Docker Compose](#docker-compose)


## What is Immich Kiosk?
I made Immich Kiosk as a lightweight (on the client) slideshow to run on kiosk devices and browsers.

## Installation
Use via [docker](#docker-compose)

## Configuration
See the file config.example.yaml for an example config file

## Changing config via browser queries
You can configure settings for individual devices through the URL. This feature is particularly useful when you need different settings for different devices, especially if the only input option available is a URL, such as with kiosk devices.

example:

`https://{URL}?refresh=60&background_blur=false&transition=none`

Thos above would set refresh to 60 seconds, turn off the background blurred image and remove all transistions for this device/browser.


## Docker Compose
```yaml
services:
  immich-kiosk:
    image: damongolding/immich-kiosk:latest
    container_name: immich-kiosk
    environment:
      DEBUG: false
    ports:
      - 3000:3000
    volumes:
      - ./config.yaml:/config.yaml
    restart: on-failure
```
## TODO
- Album



<!-- LINKS & IMAGES -->
[immich-github-url]: https://github.com/immich-app/immich
