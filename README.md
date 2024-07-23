# Immich Kiosk

<div align="center">
  <a href="https://github.com/damongolding/immich-kiosk">
    <img src="/assets/logo.svg" width="240" height="auto" alt="Immich Kiosk windmil logo" />
  </a>
</div>

**This project is not affiliated with [immich][immich-github-url]**

## Table of Contents
- [What is Immich Kiosk?](#what-is-immich-kiosk)
- [Installation](#installation)
- [Docker Compose](#docker-compose)
- [Configuration](#configuration)
- [Changing settings via URL](#changing-settings-via-url)
- [FAQ](#faq)
- [TODO](#TODO)
- [Support](#support)


## What is Immich Kiosk?
I made Immich Kiosk as a lightweight slideshow to run on kiosk devices and browsers.

![preview 1](/assets/demo_1.jpeg)
**Image shot by Damon Golding**

![preview 2](/assets/demo_2.jpeg)
**[Image shot by @insungpandora](https://unsplash.com/@insungpandora)**

### Example 1

You have a couple of spare Raspberry Pi's laying around. One hooked up to a LCD screen and the other you connect to your TV. You install a kiosk service on them (I use [DeitPi][dietpi-url]).

You want the pi connected to the LCD screen to only show images from your recent holiday, which are stored in a album on Immich. It's an older pi so you want to disable CSS transitions, also we don't want to display the time of the image.

Using this URL `http://{URL}?album={ALBUM_ID}&transtion=none&show_time=false` would achieve what we want.

On the pi connected to the TV you want to display a random image from your library. It has to be fullscreen and we want to use the fade transition

Using this URL `http://{URL}?full_screen=true&transition=fade` would achieve what we want.

### Example 2

You want to see a random picture of your child when you open a new tab in Chrome. To achieve this set the homepage URL in Chrome to `http://{URL}?person={PERSON_ID}`.


## Installation
Use via [docker](#docker-compose) 👇


## Docker Compose

> [!NOTE]
> You can use both a yaml file and environment variables but environment variables will overwrite settings from the yaml file

### When using a yaml config file
```yaml
services:
  immich-kiosk:
    image: damongolding/immich-kiosk:latest
    container_name: immich-kiosk
    environment:
      TZ: "Europe/London"
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
    environment:
      TZ: "Europe/London"
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


## Configuration
See the file config.example.yaml for an example config file

| **yaml**          | **ENV**                 | **Value**                  | **Description**                                                                            |
|-------------------|-------------------------|----------------------------|--------------------------------------------------------------------------------------------|
| immich_url        | KIOSK_IMMICH_URL        | string                     | The URL of your Immich server, e.g. `http://192.168.1.123:2283`.                           |
| immich_api_key    | KIOK_IMMICH_API_KEY     | string                     | The API for your Immich server.                                                            |
| show_time         | KIOSK_SHOW_TIME         | bool                       | Display clock.                                                                             |
| time_format       | KIOSK_TIME_FORMAT       | 12 \| 24                   | Display clock time in either 12 hour or 24 hour format. Can either be 12 or 24.            |
| show_date         | KIOSK_SHOW_DATE         | bool                       | Display the date.                                                                          |
| date_format       | KIOSK_DATE_FORMAT       | string                     | The format of the date. default is day/month/year. Any GO date string is valid.            |
| refresh           | KIOSK_REFRESH           | int                        | The amount in seconds a image will be displayed for.                                       |
| album             | KIOSK_ALBUM             | string                     | The ID of a specific album you want to display.                                            |
| person            | KIOSK_PERSON            | string                     | The ID of a specific person you want to display. Having the album set will overwrite this. |
| fill_screen       | KIOSK_FILL_SCREEN       | bool                       | Force images to be full screen. Can lead to blurriness depending on image and screen size. |
| background_blur   | KIOSK_BACKGROUND_BLUR   | bool                       | Display a blurred version of the image as a background.                                    |
| transition        | KIOSK_TRANSITION        | none \| fade \| cross-fade | Which transition to use when changing images.                                              |
| show_progress     | KIOSK_SHOW_PROGRESS     | bool                       | Display a progress bar for when image will refresh.                                        |
| show_image_time   | KIOSK_SHOW_IMAGE_TIME   | bool                       | Display image time from METADATA (if available).                                           |
| image_time_format | KIOSK_IMAGE_TIME_FORMAT | 12 \| 24                   | Display image time in either 12 hour or 24 hour format. Can either be 12 or 24.            |
| show_image_date   | KIOSK_SHOW_IMAGE_DATE   | bool                       | Display the image date from METADATA (if available).                                       |
| image_date_format | KIOSK_IMAGE_DATE_FORMAT | string                     | The format of the image date. default is day/month/year. Any GO date string is valid.      |


## Changing settings via URL
You can configure settings for individual devices through the URL. This feature is particularly useful when you need different settings for different devices, especially if the only input option available is a URL, such as with kiosk devices.

example:

`https://{URL}?refresh=120&background_blur=false&transition=none`

Thos above would set refresh to 120 seconds (2 minutes), turn off the background blurred image and remove all transitions for this device/browser.


## FAQ
- TODO


## TODO
- FAQs
- Investigate caching
- Update README images
- Homarr/Glance/other example
- HA example


## Support
If this project has been helpful to you and you wish to support me, you can do so with the button below 🙂.
[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/damongolding)



<!-- LINKS & IMAGES -->
[immich-github-url]: https://github.com/immich-app/immich
[dietpi-url]: https://dietpi.com/docs/software/desktop/#chromium
