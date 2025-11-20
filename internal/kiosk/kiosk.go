// kiosk constants
package kiosk

import "github.com/charmbracelet/lipgloss"

type Source string

const (
	GlobalCache = "kiosk-cache"

	ConfigValidationWarning string = "warning"
	ConfigValidationError   string = "error"
	ConfigValidationOff     string = "off"

	AlbumKeywordAll        string = "all"
	AlbumKeywordOwned      string = "owned"
	AlbumKeywordShared     string = "shared"
	AlbumKeywordFavourites string = "favourites"
	AlbumKeywordFavorites  string = "favorites"

	PersonKeywordAll string = "all"

	SourceAlbum     Source = "ALBUM"
	SourceDateRange Source = "DATE_RANGE_ALBUM"
	SourcePerson    Source = "PERSON"
	SourceRandom    Source = "RANDOM"
	SourceTag       Source = "TAG"
	SourceMemories  Source = "MEMORIES"

	LayoutLandscape          string = "landscape"
	LayoutPortrait           string = "portrait"
	LayoutSplitview          string = "splitview"
	LayoutSplitviewLandscape string = "splitview-landscape"

	PortraitOrientation  string = LayoutPortrait
	LandscapeOrientation string = LayoutLandscape
	SquareOrientation    string = "square"

	RedirectExternal string = "EXTERNAL"
	RedirectInternal string = "INTERNAL"

	FavoriteAlbumName = "Kiosk Favorites"

	TagSkip string = "kiosk-skip"

	LikeButtonActionFavorite string = "favorite"
	LikeButtonActionAlbum    string = "album"
	HideButtonActionTag      string = "tag"
	HideButtonActionArchive  string = "archive"

	HistoryIndicator string = "*"
	HistoryLimit     int    = 20

	ThemeFade   string = "fade"
	ThemeSolid  string = "solid"
	ThemeBubble string = "bubble"
	ThemeBlur   string = "blur"
)

var DebugID = lipgloss.NewStyle().Bold(true).Padding(0, 1).Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#1ed2bb")).Render("KIOSK")
