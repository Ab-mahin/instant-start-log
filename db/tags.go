package db

// Tag represents a row in the tags table.
type Tag struct {
	ID        int64
	MediaID   int64
	Tag       string
	CreatedAt string
}

// AddTag inserts a tag for a media item. Returns error if duplicate.
func (d *DB) AddTag(mediaID int64, tag string) error {
	_, err := d.Exec("INSERT INTO tags (media_id, tag) VALUES (?, ?)", mediaID, tag)
	return err
}

// RemoveTag deletes a specific tag from a media item.
func (d *DB) RemoveTag(mediaID int64, tag string) error {
	res, err := d.Exec("DELETE FROM tags WHERE media_id = ? AND tag = ?", mediaID, tag)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrTagNotFound
	}
	return nil
}

// ListTagsByMedia returns all tags for a given media item.
func (d *DB) ListTagsByMedia(mediaID int64) ([]Tag, error) {
	rows, err := d.Query("SELECT id, media_id, tag, created_at FROM tags WHERE media_id = ? ORDER BY tag", mediaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.MediaID, &t.Tag, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// ListAllTags returns all unique tags with their usage count.
func (d *DB) ListAllTags() (map[string]int, error) {
	rows, err := d.Query("SELECT tag, COUNT(*) FROM tags GROUP BY tag ORDER BY tag")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]int)
	for rows.Next() {
		var tag string
		var count int
		if err := rows.Scan(&tag, &count); err != nil {
			return nil, err
		}
		result[tag] = count
	}
	return result, rows.Err()
}
