package entities

import (
	"errors"

	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

var (
	ErrAttachmentNotFound  = errors.New("attachment not found")
	ErrDuplicateAttachment = errors.New("duplicate attachment")
)

type AttachmentID string

type Attachment struct {
	id         AttachmentID
	documentID DocumentID
	fileName   valueobjects.FileName
	mimeType   valueobjects.MimeType
	size       valueobjects.FileSize
	uploadedAt valueobjects.Timestamp
	expiresAt  valueobjects.ExpirationTime
}

func NewAttachment(
	id AttachmentID,
	documentID DocumentID,
	fileName valueobjects.FileName,
	mimeType valueobjects.MimeType,
	size valueobjects.FileSize,
) *Attachment {
	return &Attachment{
		id:         id,
		documentID: documentID,
		fileName:   fileName,
		mimeType:   mimeType,
		size:       size,
		uploadedAt: valueobjects.NewTimestamp(),
		expiresAt:  valueobjects.NewDefaultExpirationTime(),
	}
}

func RestoreAttachment(
	id AttachmentID,
	documentID DocumentID,
	fileName valueobjects.FileName,
	mimeType valueobjects.MimeType,
	size valueobjects.FileSize,
	uploadedAt valueobjects.Timestamp,
	expiresAt valueobjects.ExpirationTime,
) *Attachment {
	return &Attachment{
		id:         id,
		documentID: documentID,
		fileName:   fileName,
		mimeType:   mimeType,
		size:       size,
		uploadedAt: uploadedAt,
		expiresAt:  expiresAt,
	}
}

func (a *Attachment) ID() AttachmentID {
	return a.id
}

func (a *Attachment) DocumentID() DocumentID {
	return a.documentID
}

func (a *Attachment) FileName() valueobjects.FileName {
	return a.fileName
}

func (a *Attachment) MimeType() valueobjects.MimeType {
	return a.mimeType
}

func (a *Attachment) Size() valueobjects.FileSize {
	return a.size
}

func (a *Attachment) UploadedAt() valueobjects.Timestamp {
	return a.uploadedAt
}

func (a *Attachment) ExpiresAt() valueobjects.ExpirationTime {
	return a.expiresAt
}

func (a *Attachment) IsExpired() bool {
	return a.expiresAt.IsExpired()
}
