## Configuration
This section is used to generate the UnRaid template.

| **yaml**                          | **ENV**                 | **Value**                  | **Default** | **Description**                                                                            |
|-----------------------------------|-------------------------|----------------------------|-------------|--------------------------------------------------------------------------------------------|
| immich_api_key                    | KIOSK_IMMICH_API_KEY    | string                     | ""          | The API for your Immich server.                                                            |
| immich_url                        | KIOSK_IMMICH_URL        | string                     | ""          | The URL of your Immich server. MUST include a port if one is needed e.g. `http://192.168.1.123:2283`. |
| immich_external_url               | KIOSK_IMMICH_EXTERNAL_URL | string                   | ""          | The public URL of your Immich server used for generating links and QR codes.               |
| show_time                         | KIOSK_SHOW_TIME         | bool                       | false       | Display clock.                                                                             |
| time_format                       | KIOSK_TIME_FORMAT       | 24 \| 12                   | 24          | Display clock time in either 12-hour or 24-hour format. This can either be 12 or 24.       |
| show_date                         | KIOSK_SHOW_DATE         | bool                       | false       | Display the date.                                                                          |
| date_format                       | KIOSK_DATE_FORMAT       | string                     | DD/MM/YYYY  | The format of the date. default is day/month/year.                                         |
| clock_source                      | KIOSK_CLOCK_SOURCE      | client \| server           | client      | The source of the clock. Either client or server.                                          |
| duration                          | KIOSK_DURATION          | int                        | 60          | The amount in seconds an image will be displayed for.                                       |
| disable_screensaver               | KIOSK_DISABLE_SCREENSAVER | bool                     | false       | Ask browser to request a lock that prevents device screens from dimming or locking. NOTE: I haven't been able to get this to work consistently on IOS. |
| optimize_images                   | KIOSK_OPTIMIZE_IMAGES   | bool                       | false       | Whether Kiosk should resize images to match your browser screen dimensions for better performance. NOTE: In most cases this is not necessary, but if you are accessing Kiosk on a low-powered device, this may help. |
| use_gpu                           | KIOSK_USE_GPU           | bool                       | true        | Enable GPU acceleration for improved performance (e.g., CSS transforms) |
| show_archived                     | KIOSK_SHOW_ARCHIVED     | bool                       | false       | Allow assets marked as archived to be displayed.                                           |
| albums                            | KIOSK_ALBUMS            | []string                   | []          | The ID(s) of one or more albums to display. |
| album_order                       | KIOSK_ALBUM_ORDER       | random \| newest \| oldest | random      | The order an album's assets will be displayed. |
| excluded_albums                   | KIOSK_EXCLUDED_ALBUMS   | []string                   | []          | The ID(s) of a specific album or albums you want to exclude. |
| show_videos                       | KIOSK_SHOW_VIDEOS       | bool                       | false       | Enable video playback. |
| live_photos                       | KIOSK_LIVE_PHOTOS       | bool                       | false       | Enable live photos playback for albums. |
| live_photo_loop_delay             | KIOSK_LIVE_PHOTO_LOOP_DELAY | int                    | 0           | Delay in seconds before looping live photos. |
| people                            | KIOSK_PEOPLE            | []string                   | []          | The ID(s) of a specific person or people you want to display. |
| require_all_people                | KIOSK_REQUIRE_ALL_PEOPLE | bool                      | false       | Require all people to be present in an asset. |
| excluded_people                   | KIOSK_EXCLUDED_PEOPLE   | []string                   | []          | The ID(s) of a specific person or people you want to exclude. |
| dates                             | KIOSK_DATES             | []string                   | []          | A date range or ranges. |
| tags                              | KIOSK_TAGS              | []string                   | []          | Tag or tags you want to display. |
| excluded_tags                     | KIOSK_EXCLUDED_TAGS     | []string                   | []          | The tag or tags you want to exclude. |
| memories                          | KIOSK_MEMORIES          | bool                       | false       | Display memories. |
| blacklist                         | KIOSK_BLACKLIST         | []string                   | []          | The ID(s) of any specific assets you want Kiosk to skip/exclude from displaying. |
| date_filter                       | KIOSK_DATE_FILTER       | string                     | ""          | Filter person and random assets by date. |
| disable_navigation               | KIOSK_DISABLE_NAVIGATION | bool                       | false       | Disable all Kiosk's navigation (touch/click, keyboard and menu).    |
| disable_ui                        | KIOSK_DISABLE_UI        | bool                       | false       | A shortcut to set show_time, show_date, show_image_time and show_image_date to false. |
| menu_position                     | KIOSK_MENU_POSITION     | top \| bottom              | top         | Sets the position of the menu bar.    |
| frameless                         | KIOSK_FRAMELESS         | bool                       | false       | Remove borders and rounded corners on images.                                              |
| hide_cursor                       | KIOSK_HIDE_CURSOR       | bool                       | false       | Hide cursor/mouse via CSS.                                                                 |
| font_size                         | KIOSK_FONT_SIZE         | int                        | 100         | The base font size for Kiosk. Default is 100% (16px). DO NOT include the % character.      |
| background_blur                   | KIOSK_BACKGROUND_BLUR   | bool                       | true        | Display a blurred version of the image as a background.                                    |
| background_blur_amount            | KIOSK_BACKGROUND_BLUR_AMOUNT | int                   | 10          | The amount of blur to apply to the background image (sigma).                               |
| theme                             | KIOSK_THEME             | fade \| solid \| bubble    | fade        | Which theme to use. |
| layout                            | KIOSK_LAYOUT            | single \| portrait \| landscape \| splitview \| splitview-landscape | single | Which layout to use.                         |
| sleep_start                       | KIOSK_SLEEP_START       | string                     | ""          | Time (in 24hr format) to start sleep mode. |
| sleep_end                         | KIOSK_SLEEP_END         | string                     | ""          | Time (in 24hr format) to end sleep mode. |
| sleep_icon                        | KIOSK_SLEEP_ICON        | string                     | ""          | Display icon during sleep mode. |
| sleep_dim_screen                  | KIOSK_SLEEP_DIM_SCREEN  | bool                       | false       | Dim screen during sleep mode when using Fully Kiosk Browser. |
| disable_sleep                     | N/A                     | bool                       | false       | Bypass sleep mode by adding `disable_sleep=true` to the URL. |
| custom_css                        | N/A                     | bool                       | true        | Allow custom CSS to be used. |
| custom_css_class                  | N/A                     | string                     | ""          | This enables you to target specific devices in your CSS.              |
| transition                        | KIOSK_TRANSITION        | none \| fade \| cross-fade | none        | Which transition to use when changing images.                                              |
| fade_transition_duration          | KIOSK_FADE_TRANSITION_DURATION | float               | 1           | The duration of the fade (in seconds) transition.                                          |
| cross_fade_transition_duration    | KIOSK_CROSS_FADE_TRANSITION_DURATION | float         | 1           | The duration of the cross-fade (in seconds) transition.                                    |
| show_progress_bar                 | KIOSK_SHOW_PROGRESS_BAR  | bool                      | false       | Display a progress bar for when image will refresh.                                        |
| progress_bar_position             | KIOSK_PROGRESS_BAR_POSITION | top \| bottom          | top         | Sets the position of the progress bar.                                                      |
| image_fit                         | KIOSK_IMAGE_FIT         | contain \| cover \| none   | contain     | How the image should fit on the screen. Default is "contain". |
| image_effect                      | KIOSK_IMAGE_EFFECT      | none \| zoom \| smart-zoom | none        | Add an effect to images.                                                                   |
| image_effect_amount               | KIOSK_IMAGE_EFFECT_AMOUNT | int                  | 120         | Set the intensity of the image effect. Use a number between 100 (minimum) and higher, without the % symbol. |
| use_original_image                | KIOSK_USE_ORIGINAL_IMAGE | bool                      | false       | Use the original image. NOTE: If the original is not a png, gif, jpeg or webp Kiosk will fall back to using the preview. |
| show_owner                        | KIOSK_SHOW_OWNER        | bool                       | false       | Display the asset owner. Useful for shared albums.                                         |
| show_album_name                   | KIOSK_SHOW_ALBUM_NAME   | bool                       | false       | Display album names that the asset appears in.                                           |
| show_person_name                  | KIOSK_SHOW_PERSON_NAME  | bool                       | false       | Display people's names.                                                                    |
| show_person_age                   | KIOSK_SHOW_PERSON_AGE   | bool                       | false       | Display people's ages.                                                                        |
| show_image_time                   | KIOSK_SHOW_IMAGE_TIME   | bool                       | false       | Display image time from METADATA (if available).                                           |
| image_time_format                 | KIOSK_IMAGE_TIME_FORMAT | 12 \| 24                   | 24          | Display image time in either 12-hour or 24-hour format. This can either be 12 or 24.       |
| show_image_date                   | KIOSK_SHOW_IMAGE_DATE   | bool                       | false       | Display the image date from METADATA (if available).                                       |
| image_date_format                 | KIOSK_IMAGE_DATE_FORMAT | string                     | DD/MM/YYYY  | The format of the image date. default is day/month/year. |
| show_image_description            | KIOSK_SHOW_IMAGE_DESCRIPTION    | bool               | false       | Display image description from METADATA (if available).                                    |
| show_image_exif                   | KIOSK_SHOW_IMAGE_EXIF           | bool               | false       | Display image Fnumber, Shutter speed, focal length, ISO from METADATA (if available).      |
| show_image_location               | KIOSK_SHOW_IMAGE_LOCATION       | bool               | false       | Display the image location from METADATA (if available).                                   |
| show_image_qr                     | KIOSK_SHOW_IMAGE_QR             | bool               | false       | Displays a QR code linking to the original image (in Immich) next to the image metadata.   |
| hide_countries                    | KIOSK_HIDE_COUNTRIES            | []string           | []          | List of countries to hide from image_location                                              |
| show_more_info                    | KIOSK_SHOW_MORE_INFO            | bool               | true        | Enables the display of additional information about the current image(s)                   |
| show_more_info_image_link         | KIOSK_SHOW_MORE_INFO_IMAGE_LINK | bool               | true        | Shows a link to the original image (in Immich) in the additional information overlay       |
| show_more_info_qr_code            | KIOSK_SHOW_MORE_INFO_QR_CODE    | bool               | true        | Displays a QR code linking to the original image (in Immich) in the additional information overlay |
| like_button_action                | KIOSK_LIKE_BUTTON_ACTION        | []string           | [favorite]  | Action(s) to perform when the like button is clicked. Supported actions are [favorite, album]. |
| hide_button_action                | KIOSK_HIDE_BUTTON_ACTION        | []string           | [tag, archive] | Action(s) to perform when the hide button is clicked. Supported actions are [tag, archive]. |
| immich_users_api_keys             | N/A                     | map[string]string          | {}          | key:value mappings of Immich usernames to their corresponding API keys. |
| show_user                         | KIOSK_SHOW_USER         | bool                       | false       | Display the user used to fetch the image. |
| weather                           | N/A                     | []WeatherLocation          | []          | Display the current weather. |
| use_offline_mode                  | KIOSK_USE_OFFLINE_MODE  | bool                       | false       | Enable offline mode for the device. |
| offline_mode                      | N/A                     | OfflineMode{}              | {}          | Enable offline mode. |
| iframe                            | KIOSK_IFRAME            | []string                   | []          | Add iframes into Kiosk. |

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
| config_validation_level | KIOSK_CONFIG_VALIDATION_LEVEL      | error \| warning \| off | error | The level of validation to perform on the config.yaml file. |
| port                | KIOSK_PORT              | int          | 3000        | Which port Kiosk should use. NOTE: This is only typically needed when running Kiosk outside of a container. If you are running inside a container the port will need to be reflected in your compose file, e.g. `HOST_PORT:KIOSK_PORT` |
| behind_proxy        | KIOSK_BEHIND_PROXY      | bool         | false       | Is Kiosk running behind a proxy? |
| watch_config        | KIOSK_WATCH_CONFIG      | bool         | false       | Should Kiosk watch config.yaml file for changes. Reloads all connect clients if a change is detected. |
| fetched_assets_size | KIOSK_FETCHED_ASSETS_SIZE | int        | 1000        | The number of assets (data) requested from Immich per api call. min=1 max=1000. |
| http_timeout        | KIOSK_HTTP_TIMEOUT      | int          | 20          | The number of seconds before an http request will time out. |
| password            | KIOSK_PASSWORD          | string       | ""          | Please see FAQs for more info. If set, requests MUST contain the password in the GET parameters, e.g. `http://192.168.0.123:3000?password=PASSWORD`. |
| cache               | KIOSK_CACHE             | bool         | true        | Cache selective Immich api calls to reduce unnecessary calls.                              |
| prefetch            | KIOSK_PREFETCH          | bool         | true        | Pre-fetch assets in the background, so images load much quicker when duration timer ends.    |
| asset_weighting     | KIOSK_ASSET_WEIGHTING   | bool         | true        | Balances asset selection when multiple sources are used, e.g. multiple people and albums. When enabled, sources with fewer assets will show less often. |
