// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: label_images.sql

package database

import (
	"context"
	"time"
)

const assignImageToLabel = `-- name: AssignImageToLabel :exec
INSERT INTO label_images (label_id, image_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`

type AssignImageToLabelParams struct {
	LabelID int32
	ImageID int32
}

func (q *Queries) AssignImageToLabel(ctx context.Context, arg AssignImageToLabelParams) error {
	_, err := q.db.ExecContext(ctx, assignImageToLabel, arg.LabelID, arg.ImageID)
	return err
}

const getImagesForLabel = `-- name: GetImagesForLabel :many
SELECT id, registry, repository, tag, digest, created_at, updated_at, label_id, image_id
FROM images
JOIN label_images ON images.id = label_images.image_id
WHERE label_images.label_id = $1
`

type GetImagesForLabelRow struct {
	ID         int32
	Registry   string
	Repository string
	Tag        string
	Digest     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	LabelID    int32
	ImageID    int32
}

func (q *Queries) GetImagesForLabel(ctx context.Context, labelID int32) ([]GetImagesForLabelRow, error) {
	rows, err := q.db.QueryContext(ctx, getImagesForLabel, labelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetImagesForLabelRow
	for rows.Next() {
		var i GetImagesForLabelRow
		if err := rows.Scan(
			&i.ID,
			&i.Registry,
			&i.Repository,
			&i.Tag,
			&i.Digest,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LabelID,
			&i.ImageID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
