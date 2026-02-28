package clipboard

import "time"

type Content struct {
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updatedAt"`
}
