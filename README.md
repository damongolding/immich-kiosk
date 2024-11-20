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
- [Installation](#installation)
- [Docker Compose](#docker-compose)
- [Configuration](#configuration)
  - [Changing settings via URL](#changing-settings-via-url)
  - [Albums](#albums)
  - [People](#people)
  - [Image fit](#image-fit)
  - [Image effects](#image-effects)
  - [Date format](#date-format)
  - [Themes](#themes)
  - [Layouts](#layouts)
  - [Sleep mode](#sleep-mode)
  - [Cusom CSS](#custom-css)
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

![Kiosk theme fade](/assets/theme-fade.jpeg)
**Image shot by Damon Golding**

## Example 1
You have a two Raspberry Pi's. One hooked up to a LCD screen and the other you connect to your TV. You install a fullscreen browser OS or service on them (I use [DietPi][dietpi-url]).

You want the pi connected to the LCD screen to only show images from your recent holiday, which are stored in a album on Immich. It's an older pi so you want to disable CSS transitions, also we don't want to display the time of the image.

Using this URL `http://{URL}?album={ALBUM_ID}&transtion=none&show_time=false` would achieve what we want.

On the pi connected to the TV you want to display a random image from your library but only images of two specific people. We want the image to cover the whole screen (knowing some cropping will happen) and we want to use the fade transition.

Using this URL `http://{URL}?image_fit=cover&transition=fade&person=PERSON_1_ID&person=PERSON_2_ID` would achieve what we want.

## Example 2
Fanyang Meng created a digital picture frame using a Raspberry Pi Zero 2 W and Kiosk. You can read the blog post about the process [here](https://fanyangmeng.blog/build-a-selfhosted-digital-frame/).


------

## Installation
There are two main ways to install Kiosk.

### Docker (recommended)

#### *Option 1: Add Kiosk to your exsiting Immich compose stack.*

  1. Add the [kiosk service](#docker-compose) to your Immich `docker-compose.yaml` file.

  Follow from step 3 in option 2 to create the `config.yaml` file.

#### *Option 2: Create a seprate compose file for Kiosk.*

  1. Create a directory of your choice (e.g. ./immich-kiosk) to hold the `docker-compose.yaml` and config file.
     ```sh
     mkdir ./immich-kiosk
     cd ./immich-kiosk
     ```
  2. Download `docker-compose.yaml`.

     ```sh
     wget -O docker-compose.yaml https://raw.githubusercontent.com/damongolding/immich-kiosk/refs/heads/main/docker-compose.yaml
     ```

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
> Set `TZ` to a `TZ identifier` from [this list](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List)

### When using a `config.yaml` config file
```yaml
services:
  immich-kiosk:
    image: damongolding/immich-kiosk:latest
    container_name: immich-kiosk
    environment:
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
      TZ: "Europe/London"
      # Required settings
      KIOSK_IMMICH_API_KEY: "****"
      KIOSK_IMMICH_URL: "****"
      # Clock
      KIOSK_SHOW_TIME: false
      KIOSK_TIME_FORMAT: 24
      KIOSK_SHOW_DATE: false
      KIOSK_DATE_FORMAT: YYYY/MM/DD
      # Kiosk behaviour
      KIOSK_REFRESH: 60
      KIOSK_DISABLE_SCREENSAVER: false
      # Asset sources
      KIOSK_SHOW_ARCHIVED: false
      KIOSK_ALBUM: "ALBUM_ID,ALBUM_ID,ALBUM_ID"
      KIOSK_PERSON: "PERSON_ID,PERSON_ID,PERSON_ID"
      # UI
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
      KIOSK_SHOW_IMAGE_TIME: false
      KIOSK_IMAGE_TIME_FORMAT: 24
      KIOSK_SHOW_IMAGE_DATE: false
      KIOSK_IMAGE_DATE_FORMAT: YYYY-MM-DD
      KIOSK_SHOW_IMAGE_DESCRIPTION: false
      KIOSK_SHOW_IMAGE_EXIF: false
      KIOSK_SHOW_IMAGE_LOCATION: false
      KIOSK_HIDE_COUNTRIES: "HIDDEN_COUNTRY,HIDDEN_COUNTRY"
      KIOSK_SHOW_IMAGE_ID: false
      # Kiosk settings
      KIOSK_WATCH_CONFIG: false
      KIOSK_FETCHED_ASSETS_SIZE: 1000
      KIOSK_HTTP_TIMEOUT: 20
      KIOSK_PASSWORD: ""
      KIOSK_CACHE: true
      KIOSK_PREFETCH: true
      KIOSK_ASSET_WEIGHTING: true
      KIOSK_PORT: 3000
    ports:
      - 3000:3000
    restart: always
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
| disable_screensaver              | KIOSK_DISABLE_SCREENSAVER | bool                     | false       | Ask browser to request a lock that prevents device screens from dimming or locking. NOTE: I haven't been able to get this to work constantly on IOS. |
| show_archived                     | KIOSK_SHOW_ARCHIVED     | bool                       | false       | Allow assets marked as archived to be displayed.                                           |
| [album](#albums)                  | KIOSK_ALBUM             | []string                   | []          | The ID(s) of a specific album or albums you want to display. See [Albums](#albums) for more information. |
| [person](#people)                 | KIOSK_PERSON            | []string                   | []          | The ID(s) of a specific person or people you want to display. See [People](#people) for more information. |
| disable_ui                        | KIOSK_DISABLE_UI        | bool                       | false       | A shortcut to set show_time, show_date, show_image_time and image_date_format to false.    |
| frameless                         | KIOSK_FRAMELESS         | bool                       | false       | Remove borders and rounded corners on images.                                              |
| hide_cursor                       | KIOSK_HIDE_CURSOR       | bool                       | false       | Hide cursor/mouse via CSS.                                                                 |
| font_size                         | KIOSK_FONT_SIZE         | int                        | 100         | The base font size for Kiosk. Default is 100% (16px). DO NOT include the % character.      |
| background_blur                   | KIOSK_BACKGROUND_BLUR   | bool                       | true        | Display a blurred version of the image as a background.                                    |
| [theme](#themes)                  | KIOSK_THEME             | fade \| solid              | fade        | Which theme to use. See [Themes](#themes) for more information.                            |
| [layout](#layouts)                | KIOSK_LAYOUT            | single \| splitview        | single      | Which layout to use. See [Layouts](#layouts) for more information.                         |
| [sleep_start](#sleep-mode)        | KIOSK_SLEEP_START       | string                     | ""          | Time (in 24hr format) to start sleep mode. See [Sleep mode](#sleep-mode) for more information. |
| [sleep_end](#sleep-mode)          | KIOSK_SLEEP_END         | string                     | ""          | Time (in 24hr format) to end sleep mode. See [Sleep mode](#sleep-mode) for more information. |
| [custom_css](#custom-css)         | N/A                     | bool                       | true        | Allow custom CSS to be used. See [Custom CSS](#custom-css) for more information.           |
| transition                        | KIOSK_TRANSITION        | none \| fade \| cross-fade | none        | Which transition to use when changing images.                                              |
| fade_transition_duration          | KIOSK_FADE_TRANSITION_DURATION | float               | 1           | The duration of the fade (in seconds) transition.                                          |
| cross_fade_transition_duration    | KIOSK_CROSS_FADE_TRANSITION_DURATION | float         | 1           | The duration of the cross-fade (in seconds) transition.                                    |
| show_progress                     | KIOSK_SHOW_PROGRESS     | bool                       | false       | Display a progress bar for when image will refresh.                                        |
| [image_fit](#image-fit)           | KIOSK_IMAGE_FIT         | cover \| contain \| none   | contain     | How your image will fit on the screen. Default is contain. See [Image fit](#image-fit) for more info. |
| [image_effect](#image-effects)        | KIOSK_IMAGE_EFFECT        | zoom \| smart-zoom    | ""          | Add an effect to images.                                                               |
| [image_effect_amount](#image-effects) | KIOSK_IMAGE_EFFECT_AMOUNT | int                   | 120         | Set the intensity of the image effect. Use a number between 100 (minimum) and higher, without the % symbol. |
| use_original_image                | KIOSK_USE_ORIGINAL_IMAGE | bool                      | false       | Use the original image. NOTE: This will mostly likely cause kiosk to use more CPU and RAM resources. |
| show_image_time                   | KIOSK_SHOW_IMAGE_TIME   | bool                       | false       | Display image time from METADATA (if available).                                           |
| image_time_format                 | KIOSK_IMAGE_TIME_FORMAT | 12 \| 24                   | 24          | Display image time in either 12 hour or 24 hour format. Can either be 12 or 24.            |
| show_image_date                   | KIOSK_SHOW_IMAGE_DATE   | bool                       | false       | Display the image date from METADATA (if available).                                       |
| [image_date_format](#date-format) | KIOSK_IMAGE_DATE_FORMAT | string                     | DD/MM/YYYY  | The format of the image date. default is day/month/year. See [date format](#date-format) for more information. |
| show_image_description            | KIOSK_SHOW_IMAGE_DESCRIPTION | bool                  | false       | Display image description from METADATA (if available). |
| show_image_exif                   | KIOSK_SHOW_IMAGE_EXIF   | bool                       | false       | Display image Fnumber, Shutter speed, focal length, ISO from METADATA (if available).      |
| show_image_location               | KIOSK_SHOW_IMAGE_LOCATION | bool                     | false       | Display the image location from METADATA (if available).                                   |
| hide_countries                    | KIOSK_HIDE_COUNTRIES    | []string                   | []          | List of countries to hide from image_location                                                |
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
| port                | KIOSK_PORT              | int          | 3000        | Which port Kiosk should use. NOTE: that is port will need to be reflected in your compose file, e.g. `KIOSK_PORT:HOST_PORT` |
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

example:

`https://{URL}?refresh=120&background_blur=false&transition=none`

The above would set refresh to 120 seconds (2 minutes), turn off the background blurred image and remove all transitions for this device/browser.

------

## Albums

### Getting an albums ID from Immich:
1. Open Immich's web interface and click on "Albums" in the left hand navigation.
2. Click on the album you want the ID of.
3. The url will now look something like this `http://192.168.86.123:2283/albums/a04175f4-97bb-4d97-8d49-3700263043e5`.
4. The album ID is everything after `albums/`, so in this example it would be `a04175f4-97bb-4d97-8d49-3700263043e5`.

### How multiple albums work
When you specify multiple albums and/or people, Immich Kiosk creates a pool of all the requested person and album IDs.
For each image refresh, Kiosk randomly selects one ID from this pool and fetches an image associated with that album or person.

There are **three** ways you can set multiple albums:

> [!NOTE]
> These methods are applied in order of precedence. URL queries take highest priority, followed by environment variables, and finally the config.yaml file.
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

```
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

### People

### Getting a person's ID from Immich:
1. Open Immich's web interface and click on "Explore" in the left hand navigation.
2. Click on the person you want the ID of (you may have to click "view all" if you don't see them).
3. The url will now look something like this `http://192.168.86.123:2283/people/a04175f4-97bb-4d97-8d49-3700263043e5`.
4. The persons ID is everything after `people/`, so in this example it would be `a04175f4-97bb-4d97-8d49-3700263043e5`.

### How multiple people work
When you specify multiple people and/or albums, Immich Kiosk creates a pool of all the requested album and person IDs.
For each image refresh, Kiosk randomly selects one ID from this pool and fetches an image associated with that person or album.

There are **three** ways you can set multiple people ID's:

> [!NOTE]
> These methods are applied in order of precedence. URL queries take highest priority, followed by environment variables, and finally the config.yaml file.
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

```
http://{URL}?person=PERSON_ID&person=PERSON_ID&person=PERSON_ID
```
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

### Single (the default)
Display one image.

![Kiosk theme fade](/assets/theme-fade.jpeg)

### Splitview

> [!NOTE]
> Kiosk attempts to determine the orientation of each image. However, if an image lacks EXIF data,
> it may be displayed in an incorrect orientation (e.g., a portrait image shown in landscape format).

When a portrait image is fetched, Kiosk automatically retrieves a second portrait image\* and displays them side by side vertically. Landscape and square images are displayed individually.

\* If Kiosk is unable to retrieve a second unique image, the first image will be displayed individually.

![Kiosk layout splitview](/assets/layout-splitview.jpg)

### Splitview landscape

> [!NOTE]
> Kiosk attempts to determine the orientation of each image. However, if an image lacks EXIF data,
> it may be displayed in an incorrect orientation (e.g., a portrait image shown in landscape format).

When a landscape image is fetched, Kiosk automatically retrieves a second landscape image\* and displays them stacked horizontally. portrait and square images are displayed individually.

\* If Kiosk is unable to retrieve a second unique image, the first image will be displayed individually.

------

## Sleep mode

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

There is a `custom.example.css` file included that contains all the CSS selectors used by Kiosk, which you can use as a reference for your customizations.

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

| Key           | Action                        |
|---------------|-------------------------------|
| _ Spacebar    | Play/Pause and Toggle Menu    |
| → Right Arrow | Next Image(s)                 |
| ← Left Arrow  | Previous Image(s)             |

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

### Examples

```yaml
kiosk:
  redirects:
    - name: london
      url: /?weather=london

    - name: sheffield
      url: /?weather=sheffield

    - name: our-wedding
      url: /?weather=london&album=51be319b-55ea-40b0-83b7-27ac0a0d84a3

```

http://{URL}/london      -> Redirects to /?weather=london
http://{URL}/sheffield   -> Redirects to /?weather=sheffield
http://{URL}/our-wedding -> Redirects to /?weather=london&album=51be319b-55ea-40b0-83b7-27ac0a0d84a3

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
```

### Available Events

| Event           | Description                                             |
|-----------------|---------------------------------------------------------|
|`asset.new`      | Triggered when a new image is requested from Kiosk      |
|`asset.previous` | Triggered when a previous image is requested from Kiosk |
|`asset.prefetch` | Triggered when Kiosk prefecthes asset data from Immich  |
|`cache.flushed`  | Triggered when the cache is manually cleared            |

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

**Q: What is the difference between ImmichFrame and ImmichKiosk?**\
**A**:The main differences between ImmichFrame and ImmichKiosk are in how they are set up and how they interact with Immich:

- **ImmichFrame**: For individual devices
   - Installed on each device you want to use.
   - The device connects directly to Immich.
   - Data is processed on the device itself.

- **ImmichKiosk**: For multiple devices
   - Installed once on a central server.
   - Devices connect to it via a web browser, and it connects to Immich.
   - Data is processed by the Kiosk server.

In short, ImmichFrame is a 'one device, one installation, direct connection' setup, while ImmichKiosk is 'one installation, multiple devices, indirect connection.'"

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
- [ ] Add sleep mode indicator
- [ ] Whitelist for people and albums
- [ ] Exclude list
- [ ] PWA (✔ basic implimetion)
- [x] prev/next navigation
- [x] Splitview
- [ ] Splitview related images
- [ ] Docker/immich healthcheck?
- [x] Multi location weather
- [ ] Default weather location
- [ ] Redirect/friendly urls
- [ ] Webhooks

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

<!-- LINKS & IMAGES -->
[immich-github-url]: https://github.com/immich-app/immich
[dietpi-url]: https://dietpi.com/docs/software/desktop/#chromium
