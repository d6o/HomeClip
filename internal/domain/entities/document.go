package entities

import (
	"errors"

	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrInvalidDocument  = errors.New("invalid document")
)

type DocumentID string

const DefaultDocumentID = DocumentID("default")

type Document struct {
	id          DocumentID
	content     valueobjects.Content
	attachments map[AttachmentID]*Attachment
	lastUpdated valueobjects.Timestamp
	expiresAt   valueobjects.ExpirationTime
	version     int
}

func NewDocument(id DocumentID) *Document {
	return &Document{
		id:          id,
		content:     valueobjects.EmptyContent(),
		attachments: make(map[AttachmentID]*Attachment),
		lastUpdated: valueobjects.NewTimestamp(),
		expiresAt:   valueobjects.NewDefaultExpirationTime(),
		version:     1,
	}
}

func RestoreDocument(id DocumentID, content valueobjects.Content, attachments map[AttachmentID]*Attachment, lastUpdated valueobjects.Timestamp, expiresAt valueobjects.ExpirationTime, version int) *Document {
	if attachments == nil {
		attachments = make(map[AttachmentID]*Attachment)
	}
	return &Document{
		id:          id,
		content:     content,
		attachments: attachments,
		lastUpdated: lastUpdated,
		expiresAt:   expiresAt,
		version:     version,
	}
}

func (d *Document) ID() DocumentID {
	return d.id
}

func (d *Document) Content() valueobjects.Content {
	return d.content
}

func (d *Document) LastUpdated() valueobjects.Timestamp {
	return d.lastUpdated
}

func (d *Document) Version() int {
	return d.version
}

func (d *Document) UpdateContent(newContent valueobjects.Content) error {
	if d.IsExpired() {
		return valueobjects.ErrExpired
	}
	
	if d.content.Equals(newContent) {
		return nil
	}

	d.content = newContent
	d.lastUpdated = valueobjects.NewTimestamp()
	d.expiresAt = valueobjects.NewDefaultExpirationTime() // Reset expiration on update
	d.version++
	
	return nil
}

func (d *Document) Clear() {
	d.content = valueobjects.EmptyContent()
	d.lastUpdated = valueobjects.NewTimestamp()
	d.expiresAt = valueobjects.NewDefaultExpirationTime()
	d.version++
}

func (d *Document) ExpiresAt() valueobjects.ExpirationTime {
	return d.expiresAt
}

func (d *Document) IsExpired() bool {
	return d.expiresAt.IsExpired()
}

func (d *Document) AddAttachment(attachment *Attachment) error {
	if d.IsExpired() {
		return valueobjects.ErrExpired
	}
	
	if _, exists := d.attachments[attachment.ID()]; exists {
		return ErrDuplicateAttachment
	}
	
	d.attachments[attachment.ID()] = attachment
	d.lastUpdated = valueobjects.NewTimestamp()
	d.expiresAt = valueobjects.NewDefaultExpirationTime() // Reset expiration on attachment
	d.version++
	return nil
}

func (d *Document) RemoveAttachment(attachmentID AttachmentID) error {
	if _, exists := d.attachments[attachmentID]; !exists {
		return ErrAttachmentNotFound
	}
	
	delete(d.attachments, attachmentID)
	d.lastUpdated = valueobjects.NewTimestamp()
	d.version++
	return nil
}

func (d *Document) GetAttachment(attachmentID AttachmentID) (*Attachment, error) {
	attachment, exists := d.attachments[attachmentID]
	if !exists {
		return nil, ErrAttachmentNotFound
	}
	return attachment, nil
}

func (d *Document) GetAttachments() []*Attachment {
	attachments := make([]*Attachment, 0, len(d.attachments))
	for _, attachment := range d.attachments {
		attachments = append(attachments, attachment)
	}
	return attachments
}

func (d *Document) AttachmentCount() int {
	return len(d.attachments)
}