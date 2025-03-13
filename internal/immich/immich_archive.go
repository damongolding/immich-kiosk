package immich

func (a *Asset) ArchiveStatus(deviceID string, archive bool) error {

	body := UpdateAssetBody{
		IsFavorite: a.IsFavorite,
		IsArchived: archive,
	}

	return a.updateAsset(deviceID, body)
}
