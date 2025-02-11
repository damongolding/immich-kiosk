package kiosk

type Source string

const (
	AlbumKeywordAll        string = "all"
	AlbumKeywordShared     string = "shared"
	AlbumKeywordFavourites string = "favourites"
	AlbumKeywordFavorites  string = "favorites"

	SourceAlbums         Source = "ALBUM"
	SourceDateRangeAlbum Source = "DATE_RANGE_ALBUM"
	SourcePerson         Source = "PERSON"
	SourceRandom         Source = "RANDOM"
	SourceMemories       Source = "MEMORIES"

	LayoutLandscape          string = "landscape"
	LayoutPortrait           string = "portrait"
	LayoutSplitview          string = "splitview"
	LayoutSplitviewLandscape string = "splitview-landscape"

	TagSkip string = "kiosk-skip"
)
