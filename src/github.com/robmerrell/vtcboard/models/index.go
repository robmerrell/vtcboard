package models

// Index adds indexes to the database
func Index(conn *MgoConnection) error {
	prices := mainConnection.DB.C(priceCollection)
	if err := prices.EnsureIndexKey("generatedAt"); err != nil {
		return err
	}

	posts := mainConnection.DB.C(postCollection)
	if err := posts.EnsureIndexKey("uniqueId"); err != nil {
		return err
	}
	if err := posts.EnsureIndexKey("source"); err != nil {
		return err
	}
	if err := posts.EnsureIndexKey("publishedAt"); err != nil {
		return err
	}

	return nil
}
