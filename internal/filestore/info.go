package filestore

import "time"

type Info struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploadedAt"`
}
