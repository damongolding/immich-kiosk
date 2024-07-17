# Immich Kiosk

**This project is not affiliated with [immich][immich-github-url]**

## Table of Contents
- [What is Immich Kiosk?](#what-is-immich-kiosk)
- [Installation](#installation)
- [Configuration](#configuration)
- [Docker Compose](#docker-compose)


## What is Immich Kiosk?
I made Immich Kiosk as a lightweight (on the client) slideshow to run on kiosk devices and browsers.

### Example 1

You have a couple of spare Raspberry Pi's laying around. One hooked up to a LCD screen and the other you connect to your TV. You install a kiosk service on them (I use [DeitPi](https://dietpi.com/docs/software/desktop/#chromium)).

You want the pi connected to the LCD screen to only show images from your recent holiday, which are stored in a album on Immich. It's an older pi so you want to disable CSS transitions, also we don't want to display the time of the image.

Using this URL `http://{URL}?album={ALBUM_ID}&transtion=none&show_time=false` would achieve what we want.

On the pi connected to the TV you want to display a random image from your library. It has to be fullscreen and we want to use the fade transition

Using this URL `http://{URL}?full_screen=true&transition=fade` would achieve what we want.

### Example 2

You want to see a random picture of your child when you open a new tab in Chrome. To achieve this set the homepage URL in Chrome to `http://{URL}?person={PERSON_ID}`.


## Installation
Use via [docker](#docker-compose)



## Configuration
See the file config.example.yaml for an example config file

| **yaml**       | **ENV**             | **Value** | **description**                                                                            |
|----------------|---------------------|-----------|--------------------------------------------------------------------------------------------|
| immich_url     | KIOSK_IMMICH_URL    | string    | The URL of your Immich server                                                              |
| immich_api_key | KIOK_IMMICH_API_KEY | string    | The API for you Immich server                                                              |
| refresh        | KIOSK_REFRESH       | int       | The amount in seconds a image will be displayed for                                        |
| album          | KIOSK_ALBUM         | string    | The ID of a specific album you want to display                                             |
| person         | KIOSK_PERSON        | string    | The ID of a specific person you want to display. Having the album set will overwride this  |
| fill_screen    | KIOSK_FILL_SCREEN   | bool      | Force images to be full screen. Can lead to blurriness depending on image and screen size. |
| show_date      | KIOSK_SHOW_DATE     | bool      | Display the image date                                                                     |
| date_format    | KIOSK_DATE_FORMAT   | string    | The format of the date. default is day/month/year.                                         |
| show_time      | KIOSK_SHOW_TIME     | bool      | Display the image timestamp                                                                |
| time_format    | KIOSK_TIME_FORMAT   | 12 \| 24  | Display time in either 12 hour or 24 hour format.Can either be 12 or 24.                   |

## Changing config via browser queries
You can configure settings for individual devices through the URL. This feature is particularly useful when you need different settings for different devices, especially if the only input option available is a URL, such as with kiosk devices.

example:

`https://{URL}?refresh=60&background_blur=false&transition=none`

Thos above would set refresh to 60 seconds, turn off the background blurred image and remove all transitions for this device/browser.


## Docker Compose

> [!NOTE]
> You can use both a yaml file and environment variables but environment variables will overwrite settings from the yaml file

### When using a yaml config file
```yaml
services:
  immich-kiosk:
    image: damongolding/immich-kiosk:latest
    container_name: immich-kiosk
    volumes:
      - ./config.yaml:/config.yaml
    restart: on-failure
    ports:
      - 3000:3000
```

### When using environment variables
```yaml
services:
  immich-kiosk:
    image: damongolding/immich-kiosk:latest
    container_name: immich-kiosk
    enviroment:
      KIOSK_IMMICH_API_KEY: ""
      KIOSK_IMMICH_URL: ""
      KIOSK_REFRESH: 60
      KIOSK_ALBUM: ""
      KIOSK_PERSON: ""
      KIOSK_FILL_SCREEN: TRUE
      KIOSK_SHOW_DATE: TRUE
      KIOSK_DATE_FORMAT: 02/01/2006
      KIOSK_SHOW_TIME: TRUE
      KIOSK_TIME_FORMAT: 12
      KIOSK_BACKGROUND_BLUR: TRUE
      KIOSK_TRANSITION: NONE
    ports:
      - 3000:3000
    restart: on-failure
```

## TODO
- Album



<!-- LINKS & IMAGES -->
[immich-github-url]: https://github.com/immich-app/immich
