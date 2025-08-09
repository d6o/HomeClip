package application

import (
	"context"

	"github.com/d6o/homeclip/internal/application/commands"
	"github.com/d6o/homeclip/internal/application/dtos"
	"github.com/d6o/homeclip/internal/application/queries"
	"github.com/d6o/homeclip/internal/domain/entities"
)

//go:generate go tool mockgen -source=interfaces.go -destination=mocks.go -package=application

// Command Handlers
type UpdateContentCommandHandler interface {
	Handle(ctx context.Context, cmd commands.UpdateContentCommand) (*dtos.UpdateContentResponse, error)
}

type UploadFileCommandHandler interface {
	Handle(ctx context.Context, cmd commands.UploadFileCommand) (*entities.Attachment, error)
}

type DeleteFileCommandHandler interface {
	Handle(ctx context.Context, cmd commands.DeleteFileCommand) error
}

// Query Handlers
type GetContentQueryHandler interface {
	Handle(ctx context.Context, query queries.GetContentQuery) (*dtos.GetContentResponse, error)
}

type GetFileQueryHandler interface {
	Handle(ctx context.Context, query queries.GetFileQuery) (*queries.FileResult, error)
}

type ListFilesQueryHandler interface {
	Handle(ctx context.Context, query queries.ListFilesQuery) ([]*entities.Attachment, error)
}

// Application Service
type DocumentApplicationService interface {
	GetContent(ctx context.Context, documentID string) (*dtos.GetContentResponse, error)
	UpdateContent(ctx context.Context, documentID string, content string) (*dtos.UpdateContentResponse, error)
}