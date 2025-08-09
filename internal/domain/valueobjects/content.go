package valueobjects

import (
	"errors"
	"time"
)

const MaxContentLength = 1048576 // 1MB

var (
	ErrContentTooLarge = errors.New("content exceeds maximum allowed size")
	ErrInvalidContent  = errors.New("invalid content")
)

type Content struct {
	value string
}

func NewContent(value string) (Content, error) {
	if len(value) > MaxContentLength {
		return Content{}, ErrContentTooLarge
	}
	return Content{value: value}, nil
}

func EmptyContent() Content {
	return Content{value: ""}
}

func (c Content) Value() string {
	return c.value
}

func (c Content) Length() int {
	return len(c.value)
}

func (c Content) IsEmpty() bool {
	return c.value == ""
}

func (c Content) Equals(other Content) bool {
	return c.value == other.value
}

type Timestamp struct {
	value time.Time
}

func NewTimestamp() Timestamp {
	return Timestamp{value: time.Now().UTC()}
}

func TimestampFrom(t time.Time) Timestamp {
	return Timestamp{value: t.UTC()}
}

func (t Timestamp) Value() time.Time {
	return t.value
}

func (t Timestamp) String() string {
	return t.value.Format(time.RFC3339)
}

func (t Timestamp) Before(other Timestamp) bool {
	return t.value.Before(other.value)
}

func (t Timestamp) After(other Timestamp) bool {
	return t.value.After(other.value)
}

func (t Timestamp) IsZero() bool {
	return t.value.IsZero()
}