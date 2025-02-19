package kiosk

type Source string

const (
	AlbumKeywordAll        string = "all"
	AlbumKeywordShared     string = "shared"
	AlbumKeywordFavourites string = "favourites"
	AlbumKeywordFavorites  string = "favorites"

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

	TagSkip string = "kiosk-skip"
)
