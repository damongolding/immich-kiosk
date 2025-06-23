package immich

func (a *Asset) ArchiveStatus(deviceID string, archive bool) error {

	body := UpdateAssetBody{
		IsFavorite: a.IsFavorite,
		IsArchived: archive,
		Visibility: "timeline",
	}

	if archive {
		body.Visibility = "archive"
	}

	return a.updateAsset(deviceID, body)
}
