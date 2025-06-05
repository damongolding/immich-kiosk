package kiosk

type Source string

const (
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
)
