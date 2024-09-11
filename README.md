> [!IMPORTANT]
> An update regarding Kiosks development [here](https://github.com/damongolding/immich-kiosk/discussions/31)


# Immich Kiosk

<div align="center">
  <a href="https://github.com/damongolding/immich-kiosk">
    <img src="/assets/logo.svg" width="240" height="auto" alt="Immich Kiosk windmil logo" />
  </a>
</div>
<br />
<br />
<div align="center" style="display: flex; gap: 2rem;">

  <a href="https://github.com/damongolding/immich-kiosk/releases/latest" target="_blank" style="underline: none !important">
   <img alt="Kiosk latest release number" src="https://badgen.net/github/release/damongolding/immich-kiosk/stable">
  </a>

  <img alt="Docker pulls" src="https://badgen.net/docker/pulls/damongolding/immich-kiosk">

  <br />
  
  <img alt="GitHub Actions Workflow Status" src="https://img.shields.io/github/actions/workflow/status/damongolding/immich-kiosk/go-test.yml?label=test&color=029356">

  <img alt="GitHub Actions Workflow Status" src="https://img.shields.io/github/actions/workflow/status/damongolding/immich-kiosk/docker-release.yml?color=029356">

  <img alt="GitHub License" src="https://img.shields.io/github/license/damongolding/immich-kiosk?color=E6308A">

  <br />
  <br />

   <a href="https://www.buymeacoffee.com/damongolding" target="_blank" style="underline: none !important">
    <img src="https://cdn.buymeacoffee.com/buttons/v2/arial-yellow.png" alt="Buy Me A Coffee and support Kiosk" style="height: 46.88px !important;width: 167px !important;">
  </a>


  
</div>
<br />
<br />

> [!IMPORTANT]
> **This project is not affiliated with [Immich][immich-github-url]**

> [!WARNING]
> Like the Immich project, this project is currently in beta and may experience breaking changes.

## Table of Contents
- [What is Immich Kiosk?](#what-is-immich-kiosk)
  - [Example 1: Raspberry Pi](#example-2)
- [Installation](#installation)
- [Docker Compose](#docker-compose)
- [Configuration](#configuration)
  - [Changing settings via URL](#changing-settings-via-url)
  - [Image fit](#image-fit)
  - [Date format](#date-format)
- [Home Assistant](#home-assistant)
- [FAQ](#faq)
- [TODO / Roadmap](#todo--roadmap)
- [Support](#support)

## What is Immich Kiosk?
Immich Kiosk is a lightweight slideshow for running on kiosk devices and browsers that uses [Immich][immich-github-url] as a data source.

![preview 1](/assets/demo_1.jpg)
**Image shot by Damon Golding**

![preview 2](/assets/demo_2.jpg)
**[Image shot by @insungpandora](https://unsplash.com/@insungpandora)**


### Example 1
You have a two spare Raspberry Pi's laying around. One hooked up to a LCD screen and the other you connect to your TV. You install a fullscreen browser OS or service on them (I use [DietPi][dietpi-url]).

You want the pi connected to the LCD screen to only show images from your recent holiday, which are stored in a album on Immich. It's an older pi so you want to disable CSS transitions, also we don't want to display the time of the image.

Using this URL `http://{URL}?album={ALBUM_ID}&transtion=none&show_time=false` would achieve what we want.

On the pi connected to the TV you want to display a random image from your library but only images of two specific people. We want the image to cover the whole screen (knowing some cropping will happen) and we want to use the fade transition.

Using this URL `http://{URL}?image_fit=cover&transition=fade&person=PERSON_1_ID&person=PERSON_2_ID` would achieve what we want.


------

## Installation
Use via [docker](#docker-compose) 👇

------

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
      # Required settings
      KIOSK_IMMICH_API_KEY: "****"
      KIOSK_IMMICH_URL: "****"
      # Clock
      KIOSK_SHOW_TIME: FALSE
      KIOSK_TIME_FORMAT: 24
      KIOSK_SHOW_DATE: FALSE
      KIOSK_DATE_FORMAT: YYYY/MM/DD
      # Kiosk behaviour
      KIOSK_REFRESH: 60
      KIOSK_DISABLE_SCREENSAVER: FALSE
      # Asset sources
      KIOSK_SHOW_ARCHIVED: FALSE
      KIOSK_ALBUM: "ALBUM_ID,ALBUM_ID,ALBUM_ID"
      KIOSK_PERSON: "PERSON_ID,PERSON_ID,PERSON_ID"
      # UI
      KIOSK_DISABLE_UI: FALSE
      KIOSK_HIDE_CURSOR: FALSE
      KIOSK_BACKGROUND_BLUR: TRUE
      KIOSK_TRANSITION: NONE
      # Image display settings
      KIOSK_SHOW_PROGRESS: FALSE
      KIOSK_IMAGE_FIT: CONTAIN
      # Image metadata
      KIOSK_SHOW_IMAGE_TIME: FALSE
      KIOSK_IMAGE_TIME_FORMAT: 24
      KIOSK_SHOW_IMAGE_DATE: FALSE
      KIOSK_IMAGE_DATE_FORMAT: YYYY-MM-DD
      KIOSK_SHOW_IMAGE_EXIF: FALSE
      KIOSK_SHOW_IMAGE_LOCATION: FALSE
      # Kiosk settings
      KIOSK_PASSWORD: "****"
      KIOSK_CACHE: TRUE
    ports:
      - 3000:3000
    restart: on-failure
```

------

## Configuration
See the file config.example.yaml for an example config file

| **yaml**                          | **ENV**                 | **Value**                  | **Default** | **Description**                                                                            |
|-----------------------------------|-------------------------|----------------------------|-------------|--------------------------------------------------------------------------------------------|
| immich_url                        | KIOSK_IMMICH_URL        | string                     | ""          | The URL of your Immich server. MUST include a port if one is needed e.g. `http://192.168.1.123:2283`. |
| immich_api_key                    | KIOSK_IMMICH_API_KEY    | string                     | ""          | The API for your Immich server.                                                            |
| show_time                         | KIOSK_SHOW_TIME         | bool                       | false       | Display clock.                                                                             |
| time_format                       | KIOSK_TIME_FORMAT       | 12 \| 24                   | 24          | Display clock time in either 12 hour or 24 hour format. Can either be 12 or 24.            |
| show_date                         | KIOSK_SHOW_DATE         | bool                       | false       | Display the date.                                                                          |
| [date_format](#date-format)       | KIOSK_DATE_FORMAT       | string                     | DD/MM/YYYY  | The format of the date. default is day/month/year. See [date format](#date-format) for more information.|
| refresh                           | KIOSK_REFRESH           | int                        | 60          | The amount in seconds a image will be displayed for.                                       |
| disable_screensaver               | KIOSK_DISABLE_SCREENSAVER | bool                     | false       | Ask browser to request a lock that prevents device screens from dimming or locking.        |
| show_archived                     | KIOSK_SHOW_ARCHIVED     | bool                       | false       | Allow assets marked as archived to be displayed.                                           |
| album                             | KIOSK_ALBUM             | []string                   | []          | The ID(s) of a specific album or albums you want to display. See [FAQ: How do I set multiple albums?](#faq) to see how to implement this.|
| person                            | KIOSK_PERSON            | []string                   | []          | The ID(s) of a specific person or people you want to display. See [FAQ: How do I set multiple people?](#faq) to see how to implement this.|
| disable_ui                        | KIOSK_DISABLE_UI        | bool                       | false       | A shortcut to set show_time, show_date, show_image_time and image_date_format to false.    |
| hide_cursor                       | KIOSK_HIDE_CURSOR       | bool                       | false       | Hide cursor/mouse via CSS.                                                                 |
| background_blur                   | KIOSK_BACKGROUND_BLUR   | bool                       | true        | Display a blurred version of the image as a background.                                    |
| transition                        | KIOSK_TRANSITION        | none \| fade \| cross-fade | none        | Which transition to use when changing images.                                              |
| show_progress                     | KIOSK_SHOW_PROGRESS     | bool                       | false       | Display a progress bar for when image will refresh.                                        |
| [image_fit](#image-fit)           | KIOSK_IMAGE_FIT         | cover \| contain \| none   | contain     | How your image will fit on the screen. Default is contain. See [Image fit](#image-fit) for more info. |
| show_image_time                   | KIOSK_SHOW_IMAGE_TIME   | bool                       | false       | Display image time from METADATA (if available).                                           |
| image_time_format                 | KIOSK_IMAGE_TIME_FORMAT | 12 \| 24                   | 24          | Display image time in either 12 hour or 24 hour format. Can either be 12 or 24.            |
| show_image_date                   | KIOSK_SHOW_IMAGE_DATE   | bool                       | false       | Display the image date from METADATA (if available).                                       |
| [image_date_format](#date-format) | KIOSK_IMAGE_DATE_FORMAT | string                     | DD/MM/YYYY  | The format of the image date. default is day/month/year. See [date format](#date-format) for more information. |
| show_image_exif                   | KIOSK_SHOW_IMAGE_EXIF   | bool                       | false       | Display image Fnumber, Shutter speed, focal length, ISO from METADATA (if available).      |
| show_image_location               | KIOSK_SHOW_IMAGE_LOCATION | bool                     | false       | Display the image location from METADATA (if available).                                   |

### Additional options
The below options are NOT configurable through URL params. In the `config.yaml` file they sit under `kiosk` (demo below and in example `config.yaml`)

```yaml
immich_url: "****"
immich_api_key: "****"

// 👇 Additional options
kiosk:
  password: "****"
  cache: true

```


| **yaml**          | **ENV**                 | **Value**                  | **Description**                                                                            |
|-------------------|-------------------------|----------------------------|--------------------------------------------------------------------------------------------|
| password          | KIOSK_PASSWORD          | string                     | Please see FAQs for more info. If set, requests MUST contain the password in the GET parameters  e.g. `http://192.168.0.123:3000?password=PASSWORD`. |
| cache             | KIOSK_CACHE             | bool                       | Cache selective Immich api calls to reduce unnecessary calls. Default is true.             |


------

## Changing settings via URL
You can configure settings for individual devices through the URL. This feature is particularly useful when you need different settings for different devices, especially if the only input option available is a URL, such as with kiosk devices.

example:

`https://{URL}?refresh=120&background_blur=false&transition=none`

Thos above would set refresh to 120 seconds (2 minutes), turn off the background blurred image and remove all transitions for this device/browser.

------

## Image fit

This controls how the image will fit on your screen.
The options are:

### Contain (the default)
The image keeps its aspect ratio, but is resized to fit the whole screen. If the image is smaller than your screen, there will be some fuzzyness to your image.

### Cover
The image will cover the whole screen. To achieve this the image will mostly likely have some clipping/cropping and if the image is smaller than your screen, there will be some fuzzyness to your image.

### None
The image is centered and displayed "as is". If the image is larger than your screen it will be scaled down to fit your screen.


------

## Date format
> [!NOTE]
> Some characters, such as `/` and `:` are not allowed in URL params.
> So while you can set the date layout via URL params, I would suggest setting them via `config.yaml` or environment variables.


You can use the below values to create your preferred date layout.

| **Value**   | **Example output**  |
|-------------|--------------|
| YYYY        | 2024         |
| YY          | 24           |
| MMMM        | August       |
| MMM         | Aug          |
| MM          | 08           |
| M           | 8            |
| DDDD        | Monday       |
| DDD         | Mon          |
| DD          | 04           |
| D           | 4            |

### Date layout examples
These examples assume that today's date is the 22nd of August 2024.

* "YYYY-MM-DD"        => "2024-08-22"
* "YYYY/MM/DD"        => "2024/08/22"
* "YYYY:MM:DD"        => "2024:08:22"
* "YYYY MM DD"        => "2024 08 22"
* "YYYY MMM (DDD)"    => "2024 Aug (Thur)"
* "DDDD DD MMMM YYYY" => "Thursday 22 August 2024"

------

## Home Assistant

> [!NOTE]
> These examples are community Kiosk implementations.
> I am unable to provide support for Home Assistant via issues.

While I did not create Kiosk with [Home Assistant](https://www.home-assistant.io) in mind. I thought it would be useful to add Kiosk implementations I have come across.

### Using Kiosk to add a slideshow in Home Assistant.

1. Open up the dahsboard you want to add the slideshow to in edit mode.
2. Hit "add card" and search "webpage".
3. Enter the your Immich Kiosk url in the URL field e.g. `http://192.168.0.123:3000`
4. If you want to have some specific settings for the slideshow you can add them to the *[URL](#changing-settings-via-url)

\* I would suggest disabling all the UI i.e. `http://192.168.0.123:3000?disable_ui=true`


### Using Immich Kiosk as an image source for Wallpanel in HomeAssistant:

```yaml
  wallpanel:
    enabled: true
    image_fit: cover
    idle_time: 10
    screensaver_entity: input_boolean.kiosk
    screensaver_stop_navigation_path: /dashboard-kiosk
    fullscreen: true
    display_time: 86400
    image_url: >-
      http://{immich-kiosk-url}/image?person=PERSON_1_ID&person=PERSON_2_ID
    cards:
      - type: vertical-stack
        cards:
          - type: custom:weather-card
            details: true
            forecast: true
            hourly_forecast: false
            name: Weather
            entity: weather.pirateweather
            current: true
            number_of_forecasts: '6'
          - type: custom:horizon-card
            darkMode: true
            showAzimuth: true
            showElevation: true
```


------

## FAQ

![no-wifi icon](/assets/offline.svg)\
**Q: What is the no wifi icon?**\
**A**: This icon shows when the front end can't connect to the back end .

**Q: Can I use this to set Immich images as my Home Assistant dashboard background?**\
**A**: Yes! Just navigate to the dashboard with the view you wish to add the image background to.
Enter edit mode and click the ✏ next to the view you want to add the image to.
Then select the "background" tab and toggle on "Local path or web URL" and enter your url with path `/image` e.g. `http://192.168.0.123:3000/image`.
If you want to specify an album or a person you can also add that to the url e.g. `http://192.168.0.123:3000/image?album=ALBUM_ID`

**Q: Do I need to a docker service for each client?**\
**A**: Nope. Just one that your client(s) will connect to.

**Q: Do I have to use port 3000?**\
**A**: Nope. Just change the host port in your docker compose file i.e. `- 3000:3000` to `- PORT_YOU_WANT:3000`

**Q: How do I get a album ID?**\
**A**: Open Immich's web interface and click on "Albums" in the left hand navigation.
Click on the album you want the ID of.
The url will now look something like this `http://192.168.86.123:2283/albums/a04175f4-97bb-4d97-8d49-3700263043e5`.
The album ID is everything after `albums/`, so in this example it would be `a04175f4-97bb-4d97-8d49-3700263043e5`.

**Q: How do I get a persons ID?**\
**A**: Open Immich's web interface and click on "Explore" in the left hand navigation.
Click on the person you want the ID of (you may have to click "view all" if you don't see them).
The url will now look something like this `http://192.168.86.123:2283/people/a04175f4-97bb-4d97-8d49-3700263043e5`.
The persons ID is everything after `people/`, so in this example it would be `a04175f4-97bb-4d97-8d49-3700263043e5`.

**Q: How do I set multiple people?**\
**A**: 👇
* via config.yaml file
```yaml
person:
  - PERSON_ID
  - PERSON_ID
```

* via ENV in your docker-compose file use a `,` to separate IDs
```yaml
environment:
  KIOSK_PERSON: "PERSON_ID,PERSON_ID,PERSON_ID"
```

* via url quires `http://{URL}?person=PERSON_ID&person=PERSON_ID&person=PERSON_ID`

**Q: How do I set multiple albums?**\
**A**: 👇
* via config.yaml file
```yaml
album:
  - ALBUM_ID
  - ALBUM_ID
```

* via ENV in your docker-compose file use a `,` to separate IDs
```yaml
environment:
  KIOSK_ALBUM: "ALBUM_ID,ALBUM_ID,ALBUM_ID"
```

* via url quires `http://{URL}?album=ALBUM_ID&album=ALBUM_ID&album=ALBUM_ID`

**Q: How do I set/use a password?**\
**A**: 👇

> [!WARNING]
> This feature is meant for edgecase senarios and offers very little in terms of protection.
> If you are aiming to expose Kiosk beyond your local network, please investigate more secure alternatives.

via config.yaml file
```yaml
kiosk:
  password: 12345
```

via ENV in your docker-compose file
```yaml
environment:
  KIOSK_PASSWORD: "12345"
```


Then to access Kiosk you MUST add the password param in your URL e.g. http://{URL}?password=12345

------

## TODO / Roadmap
- Exclude video thumbnails from being displayed
- Clock/timestamp shadow redesign
- Whitelist for people and albums
- Exclude list
- Use favourites as image pool sauce

------

## Support
If this project has been helpful to you and you wish to support me, you can do so with the button below 🙂.

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/damongolding)


<!-- LINKS & IMAGES -->
[immich-github-url]: https://github.com/immich-app/immich
[dietpi-url]: https://dietpi.com/docs/software/desktop/#chromium
