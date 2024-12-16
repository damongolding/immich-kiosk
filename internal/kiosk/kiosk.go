package kiosk

type Source string

const (
	AlbumKeywordAll        string = "all"
	AlbumKeywordShared     string = "shared"
	AlbumKeywordFavourites string = "favourites"
	AlbumKeywordFavorites  string = "favorites"

	SourceAlbums Source = "ALBUM"
	SourcePerson Source = "PERSON"
	SourceRandom Source = "RANDOM"
)
