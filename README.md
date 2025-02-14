# Immich Kiosk

<div align="center">
  <a href="https://github.com/damongolding/immich-kiosk">
    <img src="/assets/logo.svg" width="240" height="auto" alt="Immich Kiosk windmill logo" />
  </a>
</div>
<br />
<br />
<div align="center" style="display: flex; gap: 2rem;">

  [![Awesome](https://raw.githubusercontent.com/awesome-selfhosted/awesome-selfhosted/master/_static/awesome.png)](https://github.com/awesome-selfhosted/awesome-selfhosted#photo-and-video-galleries)

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
  - [Requirements](#requirements)
  - [Key features](#key-features)
  - [Example 1: Raspberry Pi](#example-1)
  - [Community example: Raspberry Pi](#example-2)
- [Installation](#installation)
- [Docker Compose](#docker-compose)
- [Android](#android)
- [Configuration](#configuration)
  - [Changing settings via URL](#changing-settings-via-url)
  - [Albums](#albums)
  - [People](#people)
  - [Date range](#date-range)
  - [Filters](#filters)
  - [Image fit](#image-fit)
  - [Image effects](#image-effects)
  - [Date format](#date-format)
  - [Themes](#themes)
  - [Layouts](#layouts)
  - [Sleep mode](#sleep-mode)
  - [Custom CSS](#custom-css)
  - [Weather](#weather)
- [Navigation Controls](#navigation-controls)
- [Redirects](#redirects)
- [PWA](#pwa)
- [Webhooks](#webhooks)
- [Home Assistant](#home-assistant)
- [FAQ](#faq)
- [TODO / Roadmap](#todo--roadmap)
- [Support](#support)
- [Help](#help)
- [Contributing](#contributing)

## What is Immich Kiosk?
Immich Kiosk is a lightweight slideshow for running on kiosk devices and browsers that uses [Immich][immich-github-url] as a data source.

## Requirements
- A reachable Immich server that is running version v1.117.0 or above.
- A browser from [this supported list](https://browserslist.dev/?q=PiAwLjIl) or higher.

## Key features
- Simple installation and updates via Docker.
- Lightweight, responsive frontend for smooth performance.
- Display random images from your Immich collection, or curate specific albums and people.
- Fully customizable appearance with flexible transitions.
- Add a live clock with adjustable formats.
- Define default settings for all devices through environment variables or YAML config files.
- Configure device-specific settings using URL parameters.

![Kiosk theme fade](/assets/preview.jpeg)
**Image shot by Damon Golding**

## Example 1
You have a two Raspberry Pi's. One hooked up to a LCD screen and the other you connect to your TV. You install a fullscreen browser OS or service on them (I use [DietPi][dietpi-url]).

You want the pi connected to the LCD screen to only show images from your recent holiday, which are stored in a album on Immich. It's an older pi so you want to disable CSS transitions, also we don't want to display the time of the image.

Using this URL `http://{URL}?album={ALBUM_ID}&transtion=none&show_time=false` would achieve what we want.

On the pi connected to the TV you want to display a random image from your library but only images of two specific people. We want the image to cover the whole screen (knowing some cropping will happen) and we want to use the fade transition.

Using this URL `http://{URL}?image_fit=cover&transition=fade&person=PERSON_1_ID&person=PERSON_2_ID` would achieve what we want.

## Example 2
Fanyang Meng created a digital picture frame using a Raspberry Pi Zero 2 W and Kiosk. You can read the blog post about the process [here](https://fanyangmeng.blog/build-a-selfhosted-digital-frame/).

This example includes instructions on how to autoboot a Raspberry Pi directly into Immich Kiosk.


------

## Installation
There are two main ways to install Kiosk: **Docker** or **Binary**.

### Docker (recommended)

#### *Option 1: Add Kiosk to your exsiting Immich compose stack.*

  1. Add the [kiosk service](#docker-compose) to your Immich `docker-compose.yaml` file.

  Follow from *step 3* in *option 2* to create the `config.yaml` file.

#### *Option 2: Create a separate compose file for Kiosk.*

  1. Create a directory of your choice (e.g. ./immich-kiosk) to hold the `docker-compose.yaml` and config file.
     ```sh
     mkdir ./immich-kiosk
     cd ./immich-kiosk
     ```
  2. Download `docker-compose.yaml`.

     ```sh
     wget -O docker-compose.yaml https://raw.githubusercontent.com/damongolding/immich-kiosk/refs/heads/main/docker-compose.yaml
     ```

     Set `Lang` to a `language code` from [this list](/assets/locales.md))
     Set `TZ` to a `TZ identifier` from [this list](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List)

  3. Create the `config.yaml` file.

     > You may use [environment variables](#when-using-environment-variables) if preferred.

     Create config dir and download `config.yaml` file.

     ```sh
     mkdir ./config
     wget -O ./config/config.yaml https://raw.githubusercontent.com/damongolding/immich-kiosk/refs/heads/main/config.example.yaml
     ```

  4. Modify `config.yaml` file.

     Only the `immich_url` and `immich_api_key` are required fields.

  5. Start the container

      ```sh
       docker compose up -d
       ```

### Binary

> [!TIP]
> Use something like `systemd` to automate starting the Kiosk binary.

1. Download the binary file
   Vist [the latest release](https://github.com/damongolding/immich-kiosk/releases/latest) and scroll to the assets at the bottom of the release notes.
   Download the archive file that matches your machines architecture and unarchive.

2. Create config dir and download `config.yaml` file.

   ```sh
   mkdir ./config
   wget -O ./config/config.yaml url
   ```

3. Modify `config.yaml` file.

   Only the `immich_url` and `immich_api_key` are required fields.

4. Start Kiosk
   ```sh
   ./kiosk
   ```


------

## Docker Compose

> [!NOTE]
> You can use both a yaml file and environment variables but environment variables will overwrite settings from the yaml file
> Set `Lang` to a `language code` from [this list](/assets/locales.md))
> Set `TZ` to a `TZ identifier` from [this list](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List)

### When using a `config.yaml` config file
```yaml
services:
  immich-kiosk:
    image: damongolding/immich-kiosk:latest
    container_name: immich-kiosk
    environment:
      LANG: "en_GB"
      TZ: "Europe/London"
    volumes:
      # Mount the directory with config.yaml inside
      - ./config:/config
    restart: always
    ports:
      - 3000:3000
```

### When using environment variables

> [!TIP]
> You do not need to specifiy all of these.
> If you want the default behaviour/value you can omit it from you compose file.

```yaml
services:
  immich-kiosk:
    image: damongolding/immich-kiosk:latest
    container_name: immich-kiosk
    environment:
      LANG: "en_GB"
      TZ: "Europe/London"
      # Required settings
      KIOSK_IMMICH_API_KEY: "****"
      KIOSK_IMMICH_URL: "****"
      # External url for image links/QR codes
      KIOSK_IMMICH_EXTERNAL_URL: ""
      # Clock
      KIOSK_SHOW_TIME: false
      KIOSK_TIME_FORMAT: 24
      KIOSK_SHOW_DATE: false
      KIOSK_DATE_FORMAT: YYYY/MM/DD
      # Kiosk behaviour
      KIOSK_REFRESH: 60
      KIOSK_DISABLE_SCREENSAVER: false
      KIOSK_OPTIMIZE_IMAGES: false
      KIOSK_USE_GPU: true
      # Asset sources
      KIOSK_SHOW_ARCHIVED: false
      KIOSK_ALBUM: "ALBUM_ID,ALBUM_ID,ALBUM_ID"
      KIOSK_ALBUM_ORDER: random
      KIOSK_EXCLUDED_ALBUMS: "ALBUM_ID,ALBUM_ID,ALBUM_ID"
      KIOSK_PERSON: "PERSON_ID,PERSON_ID,PERSON_ID"
      KIOSK_DATE: "DATE_RANGE,DATE_RANGE,DATE_RANGE"
      KIOSK_MEMORIES: false
      KIOSK_BLACKLIST: "ASSET_ID,ASSET_ID,ASSET_ID"
      # FILTER
      KIOSK_DATE_FILTER: ""
      # UI
      KIOSK_DISABLE_NAVIGATION: false
      KIOSK_DISABLE_UI: false
      KIOSK_FRAMELESS: false
      KIOSK_HIDE_CURSOR: false
      KIOSK_FONT_SIZE: 100
      KIOSK_BACKGROUND_BLUR: true
      KIOSK_THEME: fade
      KIOSK_LAYOUT: single
      # Sleep mode
      # KIOSK_SLEEP_START: 22
      # KIOSK_SLEEP_END: 7
      # Transistion options
      KIOSK_TRANSITION: none
      KIOSK_FADE_TRANSITION_DURATION: 1
      KIOSK_CROSS_FADE_TRANSITION_DURATION: 1
      # Image display settings
      KIOSK_SHOW_PROGRESS: false
      KIOSK_IMAGE_FIT: contain
      KIOSK_IMAGE_EFFECT: smart-zoom
      KIOSK_IMAGE_EFFECT_AMOUNT: 120
      KIOSK_USE_ORIGINAL_IMAGE: false
      # Image metadata
      KIOSK_SHOW_ALBUM_NAME: false
      KIOSK_SHOW_PERSON_NAME: false
      KIOSK_SHOW_IMAGE_TIME: false
      KIOSK_IMAGE_TIME_FORMAT: 24
      KIOSK_SHOW_IMAGE_DATE: false
      KIOSK_IMAGE_DATE_FORMAT: YYYY-MM-DD
      KIOSK_SHOW_IMAGE_DESCRIPTION: false
      KIOSK_SHOW_IMAGE_EXIF: false
      KIOSK_SHOW_IMAGE_LOCATION: false
      KIOSK_HIDE_COUNTRIES: "HIDDEN_COUNTRY,HIDDEN_COUNTRY"
      KIOSK_SHOW_IMAGE_ID: false
      KIOSK_SHOW_MORE_INFO: true
      KIOSK_SHOW_MORE_INFO_IMAGE_LINK: true
      KIOSK_SHOW_MORE_INFO_QR_CODE: true
      # Kiosk settings
      KIOSK_WATCH_CONFIG: false
      KIOSK_FETCHED_ASSETS_SIZE: 1000
      KIOSK_HTTP_TIMEOUT: 20
      KIOSK_PASSWORD: ""
      KIOSK_CACHE: true
      KIOSK_PREFETCH: true
      KIOSK_ASSET_WEIGHTING: true
      KIOSK_PORT: 3000
      KIOSK_SHOW_USER: false
    ports:
      - 3000:3000
    restart: always
```

### When checking `config.yaml` into a repository

It is recommended to avoid checking-in any secrets.

  1. Create a `docker-compose.env` file.

     ```sh
     touch docker-compose.env
     ```

  2. Remove secrets from `config.yaml` (and `docker-compose.yaml`) - generally `immich_url` and `immich_api_key`.

  3. Include your secrets in `docker-compose.env`.

     ```sh
     KIOSK_IMMICH_API_KEY=SECRET_KEY
     KIOSK_IMMICH_URL=SECRET_URL
     ```

  4. Update `docker-compose.yaml` to include `env_file:`.

     ```yaml
     services:
       immich-kiosk:
         env_file:
           - docker-compose.env
     ```

  5. Add `docker-compose.env` to `.gitignore`.

------

## Android

Although Kiosk doesn't have its own dedicated mobile app, the ImmichFrame team has developed a native Android application that's compatible with Kiosk. The app offers two key advantages:

1. Better performance through a lightweight WebView implementation (compared to running in a full browser)
2. The ability to use Kiosk as your Android device's screensaver

To get started, visit the [ImmichFrame documentation](https://github.com/immichFrame/ImmichFrame/blob/main/Install_Client.md#android).
After installing the app, simply launch it and enter your Kiosk URL to begin using the service.

------

## Configuration
See the file `config.example.yaml` for an example config file

| **yaml**                          | **ENV**                 | **Value**                  | **Default** | **Description**                                                                            |
|-----------------------------------|-------------------------|----------------------------|-------------|--------------------------------------------------------------------------------------------|
| immich_api_key                    | KIOSK_IMMICH_API_KEY    | string                     | ""          | The API for your Immich server.                                                            |
| immich_url                        | KIOSK_IMMICH_URL        | string                     | ""          | The URL of your Immich server. MUST include a port if one is needed e.g. `http://192.168.1.123:2283`. |
| immich_external_url             | KIOSK_IMMICH_EXTERNAL_URL | string                     | ""          | The public URL of your Immich server used for generating links and QR codes in the additional information overlay. Useful when accessing Immich through a reverse proxy or different external URL. Example: "https://photos.example.com". If not set, falls back to immich_url. |
| show_time                         | KIOSK_SHOW_TIME         | bool                       | false       | Display clock.                                                                             |
| time_format                       | KIOSK_TIME_FORMAT       | 12 \| 24                   | 24          | Display clock time in either 12 hour or 24 hour format. Can either be 12 or 24.            |
| show_date                         | KIOSK_SHOW_DATE         | bool                       | false       | Display the date.                                                                          |
| [date_format](#date-format)       | KIOSK_DATE_FORMAT       | string                     | DD/MM/YYYY  | The format of the date. default is day/month/year. See [date format](#date-format) for more information.|
| refresh                           | KIOSK_REFRESH           | int                        | 60          | The amount in seconds a image will be displayed for.                                       |
| disable_screensaver             | KIOSK_DISABLE_SCREENSAVER | bool                       | false       | Ask browser to request a lock that prevents device screens from dimming or locking. NOTE: I haven't been able to get this to work constantly on IOS. |
| optimize_images                   | KIOSK_OPTIMIZE_IMAGES   | bool                       | false       | Whether Kiosk should resize images to match your browser screen dimensions for better performance. NOTE: In most cases this is not necessary, but if you are accessing Kiosk on a low-powered device, this may help. |
| use_gpu                           | KIOSK_USE_GPU           | bool                       | true        | Enable GPU acceleration for improved performance (e.g., CSS transforms) |
| show_archived                     | KIOSK_SHOW_ARCHIVED     | bool                       | false       | Allow assets marked as archived to be displayed.                                           |
| [album](#albums)                  | KIOSK_ALBUM             | []string                   | []          | The ID(s) of a specific album or albums you want to display. See [Albums](#albums) for more information. |
| [album_order](#album-order)       | KIOSK_ALBUM_ORDER       | string                     | random      | The order an album's assets will be displayed. See [Album order](#album-order) for more information. |
| [excluded_albums](#exclude-albums) | KIOSK_EXCLUDED_ALBUMS  | []string                   | []          | The ID(s) of a specific album or albums you want to exclude. See [Exclude albums](#exclude-albums) for more information. |
| [experimental_album_video](#experimental-album-video-support) | KIOSK_EXPERIMENTAL_ALBUM_VIDEO  | bool | false | Enable experimental video playback for albums. See [experimental album video](#experimental-album-video-support) for more information. |
| [person](#people)                 | KIOSK_PERSON            | []string                   | []          | The ID(s) of a specific person or people you want to display. See [People](#people) for more information. |
| [date](#date-range)               | KIOSK_DATE              | []string                   | []          | A date range or ranges in `YYYY-MM-DD_to_YYYY-MM-DD` format. See [Date range](#date-range) for more information. |
| memories                          | KIOSK_MEMORIES          | bool                       | false       | Display memory lane assets. |
| blacklist                         | KIOSK_BLACKLIST         | []string                   | []          | The ID(s) of any specific assets you want Kiosk to skip/exclude from displaying. You can also tag assets in Immich with "kiosk-skip" to achieve the same. |
| [date_filter](#filters)           | KIOSK_DATE_FILTER       | string                     | ""          | Filter person and random assets by date. See [date filter](#filters) for more information. |
| disable_navigation               | KIOSK_DISABLE_NAVIGATION | bool                       | false       | Disable all Kiosk's navigation (touch/click, keyboard and menu).    |
| disable_ui                        | KIOSK_DISABLE_UI        | bool                       | false       | A shortcut to set show_time, show_date, show_image_time and image_date_format to false.    |
| frameless                         | KIOSK_FRAMELESS         | bool                       | false       | Remove borders and rounded corners on images.                                              |
| hide_cursor                       | KIOSK_HIDE_CURSOR       | bool                       | false       | Hide cursor/mouse via CSS.                                                                 |
| font_size                         | KIOSK_FONT_SIZE         | int                        | 100         | The base font size for Kiosk. Default is 100% (16px). DO NOT include the % character.      |
| background_blur                   | KIOSK_BACKGROUND_BLUR   | bool                       | true        | Display a blurred version of the image as a background.                                    |
| [theme](#themes)                  | KIOSK_THEME             | fade \| solid              | fade        | Which theme to use. See [Themes](#themes) for more information.                            |
| [layout](#layouts)                | KIOSK_LAYOUT            | [Layouts](#layouts)        | single      | Which layout to use. See [Layouts](#layouts) for more information.                         |
| [sleep_start](#sleep-mode)        | KIOSK_SLEEP_START       | string                     | ""          | Time (in 24hr format) to start sleep mode. See [Sleep mode](#sleep-mode) for more information. |
| [sleep_end](#sleep-mode)          | KIOSK_SLEEP_END         | string                     | ""          | Time (in 24hr format) to end sleep mode. See [Sleep mode](#sleep-mode) for more information. |
| [disable_sleep](#sleep-mode)      | N/A                     | bool                       | false       | Bypass sleep mode by adding `disable_sleep=true` to the URL. See [Sleep mode](#sleep-mode) for more information. |
| [custom_css](#custom-css)         | N/A                     | bool                       | true        | Allow custom CSS to be used. See [Custom CSS](#custom-css) for more information.           |
| transition                        | KIOSK_TRANSITION        | none \| fade \| cross-fade | none        | Which transition to use when changing images.                                              |
| fade_transition_duration          | KIOSK_FADE_TRANSITION_DURATION | float               | 1           | The duration of the fade (in seconds) transition.                                          |
| cross_fade_transition_duration    | KIOSK_CROSS_FADE_TRANSITION_DURATION | float         | 1           | The duration of the cross-fade (in seconds) transition.                                    |
| show_progress                     | KIOSK_SHOW_PROGRESS     | bool                       | false       | Display a progress bar for when image will refresh.                                        |
| [image_fit](#image-fit)           | KIOSK_IMAGE_FIT         | cover \| contain \| none   | contain     | How your image will fit on the screen. Default is contain. See [Image fit](#image-fit) for more info. |
| [image_effect](#image-effects)    | KIOSK_IMAGE_EFFECT      | zoom \| smart-zoom         | ""          | Add an effect to images.                                                                   |
| [image_effect_amount](#image-effects) | KIOSK_IMAGE_EFFECT_AMOUNT | int                  | 120         | Set the intensity of the image effect. Use a number between 100 (minimum) and higher, without the % symbol. |
| use_original_image                | KIOSK_USE_ORIGINAL_IMAGE | bool                      | false       | Use the original image. NOTE: If the original is not a png, gif, jpeg or webp Kiosk will fallback to using the preview. |
| show_album_name                   | KIOSK_SHOW_ALBUM_NAME   | bool                       | false       | Display album name(s) that the asset appears in.                                           |
| show_person_name                  | KIOSK_SHOW_PERSON_NAME  | bool                       | false       | Display person name(s).                                                                    |
| show_image_time                   | KIOSK_SHOW_IMAGE_TIME   | bool                       | false       | Display image time from METADATA (if available).                                           |
| image_time_format                 | KIOSK_IMAGE_TIME_FORMAT | 12 \| 24                   | 24          | Display image time in either 12 hour or 24 hour format. Can either be 12 or 24.            |
| show_image_date                   | KIOSK_SHOW_IMAGE_DATE   | bool                       | false       | Display the image date from METADATA (if available).                                       |
| [image_date_format](#date-format) | KIOSK_IMAGE_DATE_FORMAT | string                     | DD/MM/YYYY  | The format of the image date. default is day/month/year. See [date format](#date-format) for more information. |
| show_image_description            | KIOSK_SHOW_IMAGE_DESCRIPTION    | bool               | false       | Display image description from METADATA (if available). |
| show_image_exif                   | KIOSK_SHOW_IMAGE_EXIF           | bool               | false       | Display image Fnumber, Shutter speed, focal length, ISO from METADATA (if available).      |
| show_image_location               | KIOSK_SHOW_IMAGE_LOCATION       | bool               | false       | Display the image location from METADATA (if available).                                   |
| hide_countries                    | KIOSK_HIDE_COUNTRIES            | []string           | []          | List of countries to hide from image_location                                              |
| show_more_info                    | KIOSK_SHOW_MORE_INFO            | bool               | true        | Enables the display of additional information about the current image(s)                   |
| show_more_info_image_link         | KIOSK_SHOW_MORE_INFO_IMAGE_LINK | bool               | true        | Shows a link to the original image (in Immich) in the additional information overlay       |
| show_more_info_qr_code            | KIOSK_SHOW_MORE_INFO_QR_CODE    | bool               | true        | Displays a QR code linking to the original image (in Immich) in the additional information overlay |
| immich_users_api_keys             | N/A                     | map[string]string          | {}          | key:value mappings of Immich usernames to their corresponding API keys. See [multiple users](#multiple-users) for more information |
| show_user                         | KIOSK_SHOW_USER         | bool                       | false       | Display the user used to fetch the image. See [multiple users](#multiple-users) for more information |
| [weather](#weather)               | N/A                     | []WeatherLocation          | []          | Display the current weather. See [weather](#weather) for more information.                 |

### Additional options
The below options are NOT configurable through URL params. In the `config.yaml` file they sit under `kiosk` (demo below and in example `config.yaml`)

```yaml
immich_url: "****"
immich_api_key: "****"
// all your other config options

// 👇 Additional options
kiosk:
  password: ""
  cache: true
  prefetch: true

```


| **yaml**            | **ENV**                 | **Value**    | **Default** | **Description**                                                                            |
|---------------------|-------------------------|--------------|-------------|--------------------------------------------------------------------------------------------|
| port                | KIOSK_PORT              | int          | 3000        | Which port Kiosk should use. NOTE: This is only typically needed when running Kiosk outside of a container. If you are running inside a container the port will need to be reflected in your compose file, e.g. `HOST_PORT:KIOSK_PORT` |
| watch_config        | KIOSK_WATCH_CONFIG      | bool         | false       | Should Kiosk watch config.yaml file for changes. Reloads all connect clients if a change is detected. |
| fetched_assets_size | KIOSK_FETCHED_ASSETS_SIZE | int        | 1000        | The number of assets (data) requested from Immich per api call. min=1 max=1000. |
| http_timeout        | KIOSK_HTTP_TIMEOUT      | int          | 20          | The number of seconds before an http request will time out. |
| password            | KIOSK_PASSWORD          | string       | ""          | Please see FAQs for more info. If set, requests MUST contain the password in the GET parameters, e.g. `http://192.168.0.123:3000?password=PASSWORD`. |
| cache               | KIOSK_CACHE             | bool         | true        | Cache selective Immich api calls to reduce unnecessary calls.                              |
| prefetch            | KIOSK_PREFETCH          | bool         | true        | Pre-fetch assets in the background, so images load much quicker when refresh timer ends.    |
| asset_weighting     | KIOSK_ASSET_WEIGHTING   | bool         | true        | Balances asset selection when multiple sources are used, e.g. multiple people and albums. When enabled, sources with fewer assets will show less often. |


------

## Changing settings via URL
You can configure settings for individual devices through the URL. This feature is particularly useful when you need different settings for different devices, especially if the only input option available is a URL, such as with kiosk devices.

Example:

`https://{URL}?refresh=120&background_blur=false&transition=none`

The above would set refresh to 120 seconds (2 minutes), turn off the background blurred image and remove all transitions for this device/browser.

------

## Multiple Users

> [!TIP]
> You can remove specific asset sources that were previously set in your `config.yaml` or environment variables by using `none` in the URL query parameters.
>
> Example:
> ```url
> https://{URL}?user=USER&person=none&album=none
> ```
>
> This will disable both the 'person' and 'album' asset sources.

Immich Kiosk supports multiple user API keys. Here's how to set it up:

1. Configure multiple users by adding their API keys to the `immich_users_api_keys` field in your `config.yaml`:
```yaml
immich_users_api_keys:
  john: "api_key_here"
  jane: "api_key_here"
```

2. Access a specific user's content by including the `user` parameter in the URL:
```
https://{URL}?user=john
```

> [!NOTE]
> If no user is specified in the URL, Immich Kiosk will default to using the global API key defined in `immich_api_key`.

## Albums

### Getting an albums ID from Immich
1. Open Immich's web interface and click on "Albums" in the left hand navigation.
2. Click on the album you want the ID of.
3. The url will now look something like this `http://192.168.86.123:2283/albums/a04175f4-97bb-4d97-8d49-3700263043e5`.
4. The album ID is everything after `albums/`, so in this example it would be `a04175f4-97bb-4d97-8d49-3700263043e5`.

### How multiple albums work
When you specify multiple albums and/or people, Immich Kiosk creates a pool of all the requested person and album IDs.
For each image refresh, Kiosk randomly selects one ID from this pool and fetches an image associated with that album or person.

There are **three** ways you can set multiple albums:

> [!NOTE]
> These methods are applied in order of precedence. URL queries take the highest priority, followed by environment variables, and finally the config.yaml file.
> Each subsequent method overwrites the settings from the previous ones.

1. via config.yaml file
```yaml
album:
  - ALBUM_ID
  - ALBUM_ID
```

2. via ENV in your docker-compose file use a `,` to separate IDs
```yaml
environment:
  KIOSK_ALBUM: "ALBUM_ID,ALBUM_ID,ALBUM_ID"
```

3. via url quires:

```url
http://{URL}?album=ALBUM_ID&album=ALBUM_ID&album=ALBUM_ID
```

### Special album keywords

#### ` all `
Will use all albums.
e.g. `http://{URL}?album=all`

#### ` shared `
Will use only shared albums.
e.g. `http://{URL}?album=shared`

####  ` favorites ` or ` favourites `
Will use only favourited assets.
e.g. `http://{URL}?album=favorites` or `http://{URL}?album=favourites`

------

## Album order

> [!NOTE]
> - When using multiple albums, the order of the albums is random.
> - When using splitview layouts:
>   - Kiosk will look for a second image with matching orientation
>   - The second image shown may not be the next sequential image
>   - Priority is given to finding images with the right aspect ratio for a balanced display

This controls the order in which the assets from the selected album(s) are displayed.

The options are:

### `random` (the default)
The assets are displayed in a random order.

### `newest`, `descending` or `desc`
The newest assets are displayed first.

### `oldest`, `ascending` or `asc`
The oldest assets are displayed first.

------

## Experimental Album Video Support

> [!WARNING]
> This feature is experimental and currently only supports album videos with certain limitations:
> - Videos will autoplay with audio muted
> - the `experimental_album_video` setting must be enabled
> - the `prefetch` setting must be enabled
> - Browser codec support may vary

### Video Optimization Recommendations

For optimal playback performance, it's strongly recommended to transcode your videos. This can be configured in Immich:
```
Admin Panel -> System Settings -> Video Transcoding
```

**Recommended Settings:**
- Codec: H264 (for maximum browser compatibility)
- Target Resolution: Select the lowest acceptable resolution for your needs

### How Video Playback Works

1. **Video Selection Process:**
   - When Kiosk selects a video from an album, it first checks if the video is cached.
   - Videos are temporarily stored in the operating system's temp directory.
   - If not cached, the video downloads in the background while another asset is displayed.

2. **Cache Management:**
   - Downloaded videos are queued for display once ready.
   - Cached videos are automatically removed after 10 minutes of inactivity to conserve disk space.
   - Videos are automatically removed when Kiosk shuts down.

3. **Playback Handling:**
   Kiosk will skip to the next asset if any of these conditions occur:
   - Video codec is unsupported by the browser
   - Playback doesn't start within 5 seconds
   - Video playback completes
   - Any playback errors are detected

### Troubleshooting Tips
- Ensure your videos are transcoded to H264 format.
- Check browser compatibility with your video codecs.
- Verify that `prefetch` is enabled in your configuration.
- Monitor system storage for cached video files.

------

## Exclude albums

This feature allows you to prevent specific albums from being displayed in the slideshow, even when using broad album selection methods like `all` or `shared`.

> [!NOTE]
> Excluded albums take precedence over album selection methods. If an album is in both the selected albums and excluded albums lists, it will be excluded.

### Getting an albums ID from Immich
1. Open Immich's web interface and click on "Albums" in the left hand navigation.
2. Click on the album you want the ID of.
3. The url will now look something like this `http://192.168.86.123:2283/albums/a04175f4-97bb-4d97-8d49-3700263043e5`.
4. The album ID is everything after `albums/`, so in this example it would be `a04175f4-97bb-4d97-8d49-3700263043e5`.


There are **three** ways you can exclude albums:

> [!NOTE]
> These methods are applied in order of precedence. URL queries take the highest priority, followed by environment variables, and finally the config.yaml file.
> Each subsequent method overwrites the settings from the previous ones.

1. via config.yaml file
```yaml
excluded_albums:
  - ALBUM_ID
  - ALBUM_ID
```

2. via ENV in your docker-compose file use a `,` to separate IDs
```yaml
environment:
  KIOSK_EXCLUDED_ALBUMS: "ALBUM_ID,ALBUM_ID,ALBUM_ID"
```

3. via url quires:

> [!NOTE]
> it is `exclude_album=` and not `excluded_albums=`

```url
http://{URL}?exclude_album=ALBUM_ID&exclude_album=ALBUM_ID&exclude_album=ALBUM_ID
```

------

### People

### Getting a person's ID from Immich
1. Open Immich's web interface and click on "Explore" in the left hand navigation.
2. Click on the person you want the ID of (you may have to click "view all" if you don't see them).
3. The url will now look something like this `http://192.168.86.123:2283/people/a04175f4-97bb-4d97-8d49-3700263043e5`.
4. The persons ID is everything after `people/`, so in this example it would be `a04175f4-97bb-4d97-8d49-3700263043e5`.

### How multiple people work
When you specify multiple people and/or albums, Immich Kiosk creates a pool of all the requested album and person IDs.
For each image refresh, Kiosk randomly selects one ID from this pool and fetches an image associated with that person or album.

There are **three** ways you can set multiple people ID's:

> [!NOTE]
> These methods are applied in order of precedence. URL queries take the highest priority, followed by environment variables, and finally the config.yaml file.
> Each subsequent method overwrites the settings from the previous ones.

1. via config.yaml file

```yaml
person:
  - PERSON_ID
  - PERSON_ID
```

2. via ENV in your docker-compose file use a `,` to separate IDs

```yaml
environment:
  KIOSK_PERSON: "PERSON_ID,PERSON_ID,PERSON_ID"
```

3. via url quires

```url
http://{URL}?person=PERSON_ID&person=PERSON_ID&person=PERSON_ID
```

------

### Date range

> [!TIP]
> You can use `today` as an alias for the current date.
> e.g. `http://{URL}?date=2023-01-01_to_today`

### How date ranges work as asset buckets
Date ranges in Immich Kiosk create distinct pools (or "buckets") of assets based on their timestamps.
Unlike filters that modify existing collections, each date range defines its own independent set of assets.
When you specify multiple date ranges, Kiosk maintains separate buckets for each range and randomly selects
one bucket during image refresh to fetch an asset from.

### Allowed formats
- `YYYY-MM-DD_to_YYYY-MM-DD` e.g.2023-01-01_to_2023-02-01
- `last-XX-days` e.g. last-30-days

There are **three** ways you can set date ranges:

> [!NOTE]
> These methods are applied in order of precedence. URL queries take the highest priority, followed by environment variables, and finally the config.yaml file.
> Each subsequent method overwrites the settings from the previous ones.

1. via config.yaml file

```yaml
date:
  - 2023-01-01_to_2023-02-01
  - 2024-11-12_to_2023-11-18
  - last-30-days
```

2. via ENV in your docker-compose file use a `,` to separate IDs

```yaml
environment:
  KIOSK_DATE: "DATE_RANGE,DATE_RANGE,DATE_RANGE"
```

3. via url quires

```url
http://{URL}?date=DATE_RANGE&date=DATE_RANGE&date=DATE_RANGE
```

------

## Filters

> [!NOTE]
> Not all filters work on all asset source/buckets.

Filters allow you to filter asset buckets (people/albums/date etc.) by certain criteria.

### Date filter

> [!NOTE]
> `date_filter` only currently applies to person and random assets.

`date_filter` accepts the same values as [date range](#date-range).

examples:
`http://{URL}?person=PERSON_ID&date_filter=2023-01-01_to_2023-02-01` will only show assets of the supplied person between 2023-01-01 and 2023-02-01.

`http://{URL}?date_filter=last-30-days` will only show (random) assets from the last 30 days.

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

## Image effects

### zoom
> [!NOTE]
> [Image fit](#image-fit) is set to `cover` automatically when this effect is used.

This effect zooms in or out to add movement to your images, with the center of the image as the focal point.

### smart-zoom
> [!NOTE]
> [Image fit](#image-fit) is set to `cover` automatically when this effect is used.
> If the image has multiple faces, Kiosk calculates the center of all faces to use as the focal point.

Smart zoom works like the regular zoom but focuses on faces and includes both zooming and panning.

> [!TIP]
> To achieve a "Ken Burns" style effect change the `image_effect_amount` to somewhere between 200-400.

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

## Themes

### Fade (the default)
Soft gradient background for the clock and image metadata.

![Kiosk theme fade](/assets/theme-fade.jpeg)

### Solid
Solid background for the clock and image metadata.

![Kiosk theme solid](/assets/theme-solid.jpeg)

------

## Layouts

> [!NOTE]
> Throughout all layouts: Kiosk attempts to determine the orientation of each image. However, if an image lacks EXIF data,
> it may be displayed in an incorrect orientation (e.g., a portrait image shown in landscape format).

The following layout options determine how images are displayed:

### Single (the default)
This is the standard layout that displays one image at a time, regardless of orientation.
It works with both portrait and landscape images.

![Kiosk theme fade](/assets/theme-fade.jpeg)

### Portrait
This layout displays one portrait-oriented image at a time.

### Landscape
This layout displays one landscape-oriented image at a time.

### Splitview
When a portrait image is fetched, Kiosk automatically retrieves a second portrait image\* and displays them side by side vertically. Landscape and square images are displayed individually.

\* If Kiosk is unable to retrieve a second unique image, the first image will be displayed individually.

![Kiosk layout splitview](/assets/layout-splitview.jpg)

### Splitview landscape
When a landscape image is fetched, Kiosk automatically retrieves a second landscape image\* and displays them stacked horizontally. portrait and square images are displayed individually.

\* If Kiosk is unable to retrieve a second unique image, the first image will be displayed individually.

------

## Sleep mode

> [!TIP]
> You can add `disable_sleep=true` to your URL quires to bypass sleepmode.

### Enabling Sleep Mode:
Setting both `sleep_start` and `sleep_end` using the 24 hour format will enable sleep mode.

### During Sleep Mode:
Kiosk will display a black screen and can optionally shows a faint clock if `show_time` or `show_date` and enabled.

### Examples
- Setting `sleep_start=22` and `sleep_end=7` will enable sleep mode from 22:00 (10pm) to 07:00 (7am).
- Setting `sleep_start=1332` and `sleep_end=1508` will enable sleep mode from 13:32 (1:32pm) to 15:08 (3:08pm).

------

# Custom CSS
> [!NOTE]
> Custom CSS is applied after all other styles, allowing you to override any default styles.

> [!WARNING]
> Be cautious when using custom CSS, as it may interfere with the normal functioning of Kiosk if not implemented correctly.
> While I'm happy to help with general Kiosk issues, I may not be able to provide specific support for problems related to custom CSS implementations.

Custom CSS allows you to further customize Kiosk's appearance beyond the built-in themes and settings.

To use custom CSS:
1. Create a file named `custom.css` in the same directory as your `docker-compose.yml` file.
2. Add your custom CSS rules to this file.
3. Mount the `custom.css` file in your container by adding the following line to the `volumes` section of your `docker-compose.yml`:
```yaml
volumes:
  - ./config:/config
  - ./custom.css:/custom.css
```
4. Restart your Kiosk container for the changes to take effect.

> [!TIP]
> Ensure that the path to your `custom.css` file is relative to your `docker-compose.yml` file.

The custom CSS will apply to all devices connected to Kiosk by default.

To disable custom CSS for a specific device, add `custom_css=false` to the URL parameters e.g. `http://{URL}?custom_css=false`

------

## Weather

> [!NOTE]
> To use the weather feature, you’ll need an API key from [OpenWeatherMap](https://openweathermap.org).

> [!TIP]
> OpenWeatherMap limits API usage to 60 calls per hour.
> Since the kiosk refreshes weather data every 10 minutes, you can monitor up to 6 locations with a single API key.

### Setting Up Weather Locations

You can configure multiple locations in the `config.yaml` file, and choose which one to display using the URL query `weather=NAME`.

### Weather Location Configuration Options:

| **Value**   | **Description** |
|-------------|-----------------|
| name        | The location’s display name (used in the URL query). |
| lat         | Latitude of the location. |
| lon         | Longitude of the location. |
| api         | OpenWeatherMap API key. |
| unit        | Units of measurement (`standard`, `metric`, or `imperial`). |
| lang        | Language code for weather descriptions (see the full list [here](https://openweathermap.org/current#multi)). |
| default     | Set this location as the default (when no location is specified) |

### Example Configuration

Here’s an example of how to add London and New York to the config.yaml file. These locations would be selectable via the URL, like this:
http://{URL}?weather=london or http://{URL}?weather=new-york.

```yaml
 weather:
  - name: london
    lat: 51.5285262
    lon: -0.2663999
    api: API_KEY
    unit: metric
    lang: en
    default: true

  - name: new-york
    lat: 40.6973709
    lon: -74.1444838
    api: API_KEY
    unit: imperial
    lang: en
```
------

## Navigation Controls

You can interact with Kiosk in three ways: touch, mouse, or keyboard.

### Touch & Click Zones

Kiosk's display is divided into interactive zones:

![Interaction zones](/assets/click-zones.jpg)

1. Left Side: Previous image(s)
2. Center top: Pause/Play and Toggle Menu
3. Right Side: Next image(s)

### Keyboard Shortcuts

| Key           | Action                                                   |
|---------------|----------------------------------------------------------|
| _ Spacebar    | Play/Pause and Toggle Menu                               |
| → Right Arrow | Next Image(s)                                            |
| ← Left Arrow  | Previous Image(s)                                        |
| i Key         | Play/Pause and Toggle Menu and display more info overlay |

------

## Redirects

Redirects provide a simple way to map short, memorable paths to longer URLs.
It's particularly useful for creating friendly URLs that redirect to more
complex endpoints with query parameters.

## How they Work

### Configuration
Redirects are defined in the `config.yaml` file under the `kiosk.redirects` section:

Each redirect consists of:
- `name`: The short path that users will use
- `url`: The destination URL where users will be redirected to
- `type`: Optional field that controls URL behavior:
  - `internal`: Keeps the URL unchanged during redirection (useful for maintaining browser history)
  - `external`: Allows URL changes during redirection (default if omitted)

### Examples

```yaml
kiosk:
  redirects:
    - name: london
      url: /?weather=london

    - name: sheffield
      url: /?weather=sheffield
      type: internal

    - name: our-wedding
      url: /?weather=london&album=51be319b-55ea-40b0-83b7-27ac0a0d84a3

```

| Source URL                  | Redirects to                                                |
|-----------------------------|-------------------------------------------------------------|
| http://{URL}/london         | /?weather=london                                            |
| http://{URL}/sheffield      | http://{URL}/sheffield                                      |
| http://{URL}/our-wedding    | /?weather=london&album=51be319b-55ea-40b0-83b7-27ac0a0d84a3 |

------

## PWA

> [!NOTE]
> IOS does not allow PWA's to prevent the screen from going to sleep.
> A work around is to lauch Kiosk then enable the [guided access](https://support.apple.com/en-gb/guide/iphone/iph7fad0d10/ios) feature.

### IOS
1. Open Safari and navigate to Kiosk.
2. Tap on the share icon in Safari's navigation bar.
3. Scroll till you see "Add to Home Screen" and tap it.
4. Tap on the newly added Kiosk icon on your home screen!

------

## Webhooks

> [!TIP]
> To include the `clientName` in your webhook payload, append `client=YOUR_CLIENT_NAME` to your URL parameters.

Kiosk can notify external services about certain events using webhooks. When enabled, Kiosk will send HTTP POST requests to your specified webhook URL(s) when these events occur.

### Enabling Webhooks

Add webhook configuration to your `config.yaml`:

> [!TIP]
> You can have multiple webhooks for different urls and events.

```yaml
webhooks:
  - url: "https://your-webhook-endpoint.com"
    event: asset.new
    secret: "my_webhook_secret" # Optional secret for securing webhooks
```

When a secret is provided, Kiosk will generate a SHA-256 HMAC signature using the webhook payload and include it in the `X-Kiosk-Signature-256` header.
This allows you to verify that webhook request came from your Kiosk instance, following the same validation pattern as [GitHub webhooks](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries).

To validate webhooks on your server, you should:
1. Get the signature from the `X-Kiosk-Signature-256` header
2. Generate a HMAC hex digest using your secret and the raw request body
3. Compare the signatures using a constant-time comparison function

### Available Events

| Event                              | Description                                             |
|------------------------------------|---------------------------------------------------------|
|`asset.new`                         | Triggered when a new image is requested from Kiosk      |
|`asset.previous`                    | Triggered when a previous image is requested from Kiosk |
|`asset.prefetch`                    | Triggered when Kiosk prefecthes asset data from Immich  |
|`cache.flushed`                     | Triggered when the cache is manually cleared            |
|`user.webhook.trigger.info_overlay` | Triggered when the "trigger webhook" button is clicked in the image details overlay |

### Webhook Payload

| Field        | Type          | Description                                          |
|--------------|---------------|------------------------------------------------------|
| `event`      | string        | The type of event, e.g., "asset.new".                |
| `timestamp`  | string (ISO)  | The time the event occurred, in ISO 8601 format.     |
| `deviceID`   | string (UUID) | Unique identifier for the device.                    |
| `clientName` | string        | Name of the client device.                           |
| `assetCount` | int           | Number of assets related to the event.               |
| `assets`     | array         | Array of asset objects.                              |
| `config`     | object        | Configuration options for the application.           |
| `meta`       | object        | Metadata about the source and version of the system. |

### Example payload

```json
{
    "event": "asset.new",
    "timestamp": "2024-11-19T11:03:07Z",
    "deviceID": "ed08beb1-6de7-4592-9827-078c3ad91ae4",
    "clientName": "dining-room-pi",
    "assetCount": 1,
    "assets": [
         {
             "id": "bb4ce63b-b80d-430f-ad37-5cfe243e08b1",
             "type": "IMAGE",
             "originalMimeType": "image/jpeg",
             "localDateTime": "2013-04-06T23:45:54Z"
             /* ... other properties omitted for brevity */
         }
     ],
     "config": {
         /* ... configuration fields omitted for brevity */
     },
     "meta": {
         "source": "immich-kiosk",
         "version": "0.13.1"
     }
}
```

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

### Using Kiosk to set Immich images as Home Assistant dashboard background
1. Navigate to the dashboard with the view you wish to add the image background to.
2. Enter edit mode and click the ✏ next to the view you want to add the image to.
3. Select the "background" tab and toggle on "Local path or web URL" and enter your url* with path /image e.g. http://192.168.0.123:3000/image.

\* If you want to specify an album or a person you can also add that to the url e.g. http://192.168.0.123:3000/image?album=ALBUM_ID

### Using Immich Kiosk as an image source for Wallpanel in Home Assistant:

> [!TIP]
> The new version of wallpanel doesn't seem to grab new images if the url is not dynamic.
> Adding a wallpanel varqaible to the Kiosk url fixes this.
> An example is adding `t=${timestamp}` the url.
> Kiosk does not need or use this data but fixes the issue in newer Wallpanel versions.

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
      http://{immich-kiosk-url}/image?person=PERSON_1_ID&person=PERSON_2_ID&t=${timestamp}
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

More documentation about WallPanel here: [https://github.com/j-a-n/lovelace-wallpanel](https://github.com/j-a-n/lovelace-wallpanel)

### Using Kiosk on Google Cast Devices

Create a photo slideshow in Home Assistant and display it on any Google Cast device (like Chromecast or Nest Hub).

Follow this community guide to get started: [Immich Kiosk Mode](https://github.com/chris-burrow-apps/home-assistant-blueprints/blob/main/immich/immich_kiosk_mode.md)

------

## FAQ

![no-wifi icon](/assets/offline.svg)\
**Q: What is the no wifi icon?**\
**A**: This icon shows when the front end can't connect to the back end.

![flush cache icon](/assets/flush-cache.svg)\
**Q: What is this icon in the menu?**\
**A**: Clicking this icon tells Kiosk to delete all cached data and refresh the current device.

**Q: Can I use this to set Immich images as my Home Assistant dashboard background?**\
**A**: Yes! Just navigate to the dashboard with the view you wish to add the image background to.
Enter edit mode and click the ✏ next to the view you want to add the image to.
Then select the "background" tab and toggle on "Local path or web URL" and enter your url with path `/image` e.g. `http://192.168.0.123:3000/image`.
If you want to specify an album or a person you can also add that to the url e.g. `http://192.168.0.123:3000/image?album=ALBUM_ID`

**Q: Do I need to a docker service for each client?**\
**A**: You only need one docker service (or binary running) that your devices(s) will connect to.

**Q: Do I have to use port 3000?**\
**A**: Nope. Just change the host port in your docker compose file i.e. `- 3000:3000` to `- PORT_YOU_WANT:3000`


**Q: How do I set/use a password?**\
**A**: 👇

> [!WARNING]
> This feature is meant for edgecase scenarios and offers very little in terms of protection.
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
- [x] Sleep mode
- [x] Add sleep mode indicator
- [ ] Whitelist for people and albums
- [x] Exclude list
- [x] PWA (✔ basic implimetion)
- [x] prev/next navigation
- [x] Splitview
- [x] Splitview related images
- [ ] Docker/immich healthcheck?
- [x] Multi location weather
- [x] Default weather location
- [x] Redirect/friendly urls
- [x] Webhooks

------

## Support
If this project has been helpful to you and you wish to support me, you can do so with the button below 🙂.

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/damongolding)

------

## Help

If you have found a bug or have an issue you can submit it [here](https://github.com/damongolding/immich-kiosk/issues/new/choose).

If you'd like to chat or need some informal help, feel free to find me in the Kiosk channel on the Immich discord server.

<a href="https://discord.com/channels/979116623879368755/1293191523927851099">
  <img style="height:32px!important" src="https://img.shields.io/badge/Immich%20Kiosk-Kiosk%20Discord?style=flat&logo=discord&logoColor=%23fff&labelColor=%235865F2&color=%235865F2" alt="Discord button">
</a>


------

## Contributing

### Prerequisites
Want to help improve Immich Kiosk? Great! Here's what you'll need to get started:

First, make sure you have these tools installed on your computer:

- [Go](https://golang.org/doc/install) - The main programming language we use
- [Taskfile](https://taskfile.dev/installation/) - Helps automate common tasks
- [Node.js](https://nodejs.org/) - For running the frontend
- [pnpm](https://pnpm.io/installation) - Package manager for the frontend

Ready to contribute? Here's how:

1. Fork the repository and create a new branch for your changes
   ```sh
   git checkout -b feature/my-feature
   ```

2. Run `task install` to set up your development environment

3. Make your changes! Just remember to:
   - Follow the existing code style
   - Add tests if you're adding new features
   - Test your changes with `task test`
   - Check code quality with `task lint`

4. Commit your changes with a helpful message
   ```sh
   git commit -m "feat: description of your change"
   ```

5. Push your changes to GitHub
   ```sh
   git push origin feature/my-feature
   ```

6. Create a Pull Request to the `main` branch
   - Tell us what your changes do and why you made them
   - Link to any related issues
   - Add screenshots if you changed anything visual

### Guidelines

We try to keep things organized, so please:
- Follow [Go best practices](https://golang.org/doc/effective_go)
- Write clear commit messages following [conventional commits](https://www.conventionalcommits.org/)
- Keep changes focused and manageable in size
- Update docs if you change how things work
- Add tests for new features

Your changes will need to pass our automated checks before being merged.

Need help? We're here for you!
- Open an issue on GitHub
- Chat with us in the [Discord channel](https://discord.com/channels/979116623879368755/1293191523927851099)


<!-- LINKS & IMAGES -->
[immich-github-url]: https://github.com/immich-app/immich
[dietpi-url]: https://dietpi.com/docs/software/desktop/#chromium
