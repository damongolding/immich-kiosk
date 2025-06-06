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

## Links
- [Documentation](http://docs.immichkiosk.app)
- [Demo](http://demo.immichkiosk.app)

## What is Immich Kiosk?
Immich Kiosk is a lightweight slideshow for running on kiosk devices and browsers that uses [Immich][immich-github-url] as a data source.

> [!IMPORTANT]
> **This project is not affiliated with [Immich][immich-github-url]**

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

## Configuration
This section is used to generate the UnRaid template.

| **yaml**                          | **ENV**                 | **Value**                  | **Default** | **Description**                                                                            |
|-----------------------------------|-------------------------|----------------------------|-------------|--------------------------------------------------------------------------------------------|
| immich_api_key                    | KIOSK_IMMICH_API_KEY    | string                     | ""          | The API for your Immich server.                                                            |
| immich_url                        | KIOSK_IMMICH_URL        | string                     | ""          | The URL of your Immich server. MUST include a port if one is needed e.g. `http://192.168.1.123:2283`. |
| immich_external_url             | KIOSK_IMMICH_EXTERNAL_URL | string                     | ""          | The public URL of your Immich server used for generating links and QR codes in the additional information overlay. Useful when accessing Immich through a reverse proxy or different external URL. Example: "https://photos.example.com". If not set, falls back to immich_url. |
| show_time                         | KIOSK_SHOW_TIME         | bool                       | false       | Display clock.                                                                             |
| time_format                       | KIOSK_TIME_FORMAT       | 24 \| 12                   | 24          | Display clock time in either 12-hour or 24-hour format. This can either be 12 or 24.       |
| show_date                         | KIOSK_SHOW_DATE         | bool                       | false       | Display the date.                                                                          |
| [date_format](#date-format)       | KIOSK_DATE_FORMAT       | string                     | DD/MM/YYYY  | The format of the date. default is day/month/year. See [date format](#date-format) for more information.|
| clock_source                      | KIOSK_CLOCK_SOURCE      | client \| server           | client      | The source of the clock. Either client or server.                                          |
| refresh                           | KIOSK_REFRESH           | int                        | 60          | The amount in seconds a image will be displayed for.                                       |
| disable_screensaver             | KIOSK_DISABLE_SCREENSAVER | bool                       | false       | Ask browser to request a lock that prevents device screens from dimming or locking. NOTE: I haven't been able to get this to work constantly on IOS. |
| optimize_images                   | KIOSK_OPTIMIZE_IMAGES   | bool                       | false       | Whether Kiosk should resize images to match your browser screen dimensions for better performance. NOTE: In most cases this is not necessary, but if you are accessing Kiosk on a low-powered device, this may help. |
| use_gpu                           | KIOSK_USE_GPU           | bool                       | true        | Enable GPU acceleration for improved performance (e.g., CSS transforms) |
| show_archived                     | KIOSK_SHOW_ARCHIVED     | bool                       | false       | Allow assets marked as archived to be displayed.                                           |
| [album](#albums)                  | KIOSK_ALBUM             | []string                   | []          | The ID(s) of a specific album or albums you want to display. See [Albums](#albums) for more information. |
| [album_order](#album-order)       | KIOSK_ALBUM_ORDER       | random \| newest \| oldest | random  | The order an album's assets will be displayed. See [Album order](#album-order) for more information. |
| [excluded_albums](#exclude-albums) | KIOSK_EXCLUDED_ALBUMS  | []string                   | []          | The ID(s) of a specific album or albums you want to exclude. See [Exclude albums](#exclude-albums) for more information. |
| [experimental_album_video](#experimental-album-video-support) | KIOSK_EXPERIMENTAL_ALBUM_VIDEO  | bool | false | Enable experimental video playback for albums. See [experimental album video](#experimental-album-video-support) for more information. |
| [person](#people)                 | KIOSK_PERSON            | []string                   | []          | The ID(s) of a specific person or people you want to display. See [People](#people) for more information. |
| [require_all_people](#require-all-people) | KIOSK_REQUIRE_ALL_PEOPLE | bool              | false       | Require all people to be present in an asset. See [Require all people](#require-all-people) for more information. |
| [excluded_people](#exclude-people) | KIOSK_EXCLUDED_PEOPLE  | []string                   | []          | The ID(s) of a specific person or people you want to exclude. See [Exclude people](#exclude-people) for more information. |
| [date](#date-range)               | KIOSK_DATE              | []string                   | []          | A date range or ranges. See [Date range](#date-range) for more information. |
| [tag](#tags)                      | KIOSK_TAG               | []string                   | []          | Tag or tags you want to display. See [Tags](#tags) for more information. |
| memories                          | KIOSK_MEMORIES          | bool                       | false       | Display memories. |
| blacklist                         | KIOSK_BLACKLIST         | []string                   | []          | The ID(s) of any specific assets you want Kiosk to skip/exclude from displaying. You can also tag assets in Immich with "kiosk-skip" to achieve the same. |
| [date_filter](#filters)           | KIOSK_DATE_FILTER       | string                     | ""          | Filter person and random assets by date. See [date filter](#filters) for more information. |
| disable_navigation               | KIOSK_DISABLE_NAVIGATION | bool                       | false       | Disable all Kiosk's navigation (touch/click, keyboard and menu).    |
| disable_ui                        | KIOSK_DISABLE_UI        | bool                       | false       | A shortcut to set show_time, show_date, show_image_time and image_date_format to false.    |
| frameless                         | KIOSK_FRAMELESS         | bool                       | false       | Remove borders and rounded corners on images.                                              |
| hide_cursor                       | KIOSK_HIDE_CURSOR       | bool                       | false       | Hide cursor/mouse via CSS.                                                                 |
| font_size                         | KIOSK_FONT_SIZE         | int                        | 100         | The base font size for Kiosk. Default is 100% (16px). DO NOT include the % character.      |
| background_blur                   | KIOSK_BACKGROUND_BLUR   | bool                       | true        | Display a blurred version of the image as a background.                                    |
| background_blur_amount            | KIOSK_BACKGROUND_BLUR_AMOUNT | int                   | 10          | The amount of blur to apply to the background image (sigma).                               |
| [theme](#themes)                  | KIOSK_THEME             | fade \| solid              | fade        | Which theme to use. See [Themes](#themes) for more information.                            |
| [layout](#layouts)                | KIOSK_LAYOUT            | single \| portrait \| landscape \| splitview \| splitview-landscape | Which layout to use. See [Layouts](#layouts) for more information.                         |
| [sleep_start](#sleep-mode)        | KIOSK_SLEEP_START       | string                     | ""          | Time (in 24hr format) to start sleep mode. See [Sleep mode](#sleep-mode) for more information. |
| [sleep_end](#sleep-mode)          | KIOSK_SLEEP_END         | string                     | ""          | Time (in 24hr format) to end sleep mode. See [Sleep mode](#sleep-mode) for more information. |
| [sleep_icon](#sleep-mode)         | KIOSK_SLEEP_ICON        | string                     | ""          | Display icon during sleep mode. See [Sleep mode](#sleep-mode) for more information. |
| [sleep_dim_screen](#sleep-mode)   | KIOSK_SLEEP_DIM_SCREEN  | bool                       | false       | Dim screen during sleep mode when using Fully Kiosk Browser. See [Sleep mode](#sleep-mode) for more information. |
| [disable_sleep](#sleep-mode)      | N/A                     | bool                       | false       | Bypass sleep mode by adding `disable_sleep=true` to the URL. See [Sleep mode](#sleep-mode) for more information. |
| [custom_css](#custom-css)         | N/A                     | bool                       | true        | Allow custom CSS to be used. See [Custom CSS](#custom-css) for more information.           |
| transition                        | KIOSK_TRANSITION        | none \| fade \| cross-fade | none        | Which transition to use when changing images.                                              |
| fade_transition_duration          | KIOSK_FADE_TRANSITION_DURATION | float               | 1           | The duration of the fade (in seconds) transition.                                          |
| cross_fade_transition_duration    | KIOSK_CROSS_FADE_TRANSITION_DURATION | float         | 1           | The duration of the cross-fade (in seconds) transition.                                    |
| show_progress                     | KIOSK_SHOW_PROGRESS     | bool                       | false       | Display a progress bar for when image will refresh.                                        |
| [image_fit](#image-fit)           | KIOSK_IMAGE_FIT         | contain \| cover \| none   | contain     | How your image will fit on the screen. Default is contain. See [Image fit](#image-fit) for more info. |
| [image_effect](#image-effects)    | KIOSK_IMAGE_EFFECT      | none \| zoom \| smart-zoom | none        | Add an effect to images.                                                                   |
| [image_effect_amount](#image-effects) | KIOSK_IMAGE_EFFECT_AMOUNT | int                  | 120         | Set the intensity of the image effect. Use a number between 100 (minimum) and higher, without the % symbol. |
| use_original_image                | KIOSK_USE_ORIGINAL_IMAGE | bool                      | false       | Use the original image. NOTE: If the original is not a png, gif, jpeg or webp Kiosk will fallback to using the preview. |
| show_owner                        | KIOSK_SHOW_OWNER        | bool                       | false       | Display the asset owner. Useful for shared albums.                                         |
| show_album_name                   | KIOSK_SHOW_ALBUM_NAME   | bool                       | false       | Display album name(s) that the asset appears in.                                           |
| show_person_name                  | KIOSK_SHOW_PERSON_NAME  | bool                       | false       | Display person name(s).                                                                    |
| show_person_age                   | KIOSK_SHOW_PERSON_AGE   | bool                       | false       | Display person age.                                                                        |
| show_image_time                   | KIOSK_SHOW_IMAGE_TIME   | bool                       | false       | Display image time from METADATA (if available).                                           |
| image_time_format                 | KIOSK_IMAGE_TIME_FORMAT | 12 \| 24                   | 24          | Display image time in either 12-hour or 24-hour format. This can either be 12 or 24.            |
| show_image_date                   | KIOSK_SHOW_IMAGE_DATE   | bool                       | false       | Display the image date from METADATA (if available).                                       |
| [image_date_format](#date-format) | KIOSK_IMAGE_DATE_FORMAT | string                     | DD/MM/YYYY  | The format of the image date. default is day/month/year. See [date format](#date-format) for more information.
| show_image_description            | KIOSK_SHOW_IMAGE_DESCRIPTION    | bool               | false       | Display image description from METADATA (if available).                                    |
| show_image_exif                   | KIOSK_SHOW_IMAGE_EXIF           | bool               | false       | Display image Fnumber, Shutter speed, focal length, ISO from METADATA (if available).      |
| show_image_location               | KIOSK_SHOW_IMAGE_LOCATION       | bool               | false       | Display the image location from METADATA (if available).                                   |
| show_image_qr                     | KIOSK_SHOW_IMAGE_QR             | bool               | false       | Displays a QR code linking to the original image (in Immich) next to the image metadata.   |
| hide_countries                    | KIOSK_HIDE_COUNTRIES            | []string           | []          | List of countries to hide from image_location                                              |
| show_more_info                    | KIOSK_SHOW_MORE_INFO            | bool               | true        | Enables the display of additional information about the current image(s)                   |
| show_more_info_image_link         | KIOSK_SHOW_MORE_INFO_IMAGE_LINK | bool               | true        | Shows a link to the original image (in Immich) in the additional information overlay       |
| show_more_info_qr_code            | KIOSK_SHOW_MORE_INFO_QR_CODE    | bool               | true        | Displays a QR code linking to the original image (in Immich) in the additional information overlay |
| [like_button_action](#like-button)            | KIOSK_LIKE_BUTTON_ACTION    | []string           | [favorite]  | Action(s) to perform when the like button is clicked. Supported actions are [favorite, album]. See [like button](#like-button) for more information. |
| [hide_button_action](#hide-button)                | KIOSK_HIDE_BUTTON_ACTION        | []string           | [tag]       | Action(s) to perform when the hide button is clicked. Supported actions are [tag, archive]. See [hide button](#hide-button) for more information. |
| immich_users_api_keys             | N/A                     | map[string]string          | {}          | key:value mappings of Immich usernames to their corresponding API keys. See [multiple users](#multiple-users) for more information. |
| show_user                         | KIOSK_SHOW_USER         | bool                       | false       | Display the user used to fetch the image. See [multiple users](#multiple-users) for more information. |
| [weather](#weather)               | N/A                     | []WeatherLocation          | []          | Display the current weather. See [weather](#weather) for more information.                 |
| use_offline_mode                  | KIOSK_USE_OFFLINE_MODE  | bool                       | false       | Enable offline mode for the device. See [offline mode](#offline-mode) for more information. |
| [offline_mode](#offline-mode)     | N/A                     | OfflineMode{}              | {}          | Enable offline mode. See [offline mode](#offline-mode) for more information. |
| [iframe](#iframe)                 | KIOSK_IFRAME            | []string                   | []          | Add iframes into Kiosk. See [iframe](#iframe) for more information. |

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
| behind_proxy        | KIOSK_BEHIND_PROXY      | bool         | false       | Is Kiosk running behind a proxy? |
| watch_config        | KIOSK_WATCH_CONFIG      | bool         | false       | Should Kiosk watch config.yaml file for changes. Reloads all connect clients if a change is detected. |
| fetched_assets_size | KIOSK_FETCHED_ASSETS_SIZE | int        | 1000        | The number of assets (data) requested from Immich per api call. min=1 max=1000. |
| http_timeout        | KIOSK_HTTP_TIMEOUT      | int          | 20          | The number of seconds before an http request will time out. |
| password            | KIOSK_PASSWORD          | string       | ""          | Please see FAQs for more info. If set, requests MUST contain the password in the GET parameters, e.g. `http://192.168.0.123:3000?password=PASSWORD`. |
| cache               | KIOSK_CACHE             | bool         | true        | Cache selective Immich api calls to reduce unnecessary calls.                              |
| prefetch            | KIOSK_PREFETCH          | bool         | true        | Pre-fetch assets in the background, so images load much quicker when refresh timer ends.    |
| asset_weighting     | KIOSK_ASSET_WEIGHTING   | bool         | true        | Balances asset selection when multiple sources are used, e.g. multiple people and albums. When enabled, sources with fewer assets will show less often. |


------

<!-- LINKS & IMAGES -->
[immich-github-url]: https://github.com/immich-app/immich
[dietpi-url]: https://dietpi.com/docs/software/desktop/#chromium
