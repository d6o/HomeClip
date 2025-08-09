package services

import (
	"context"

	"github.com/d6o/homeclip/internal/application"
	"github.com/d6o/homeclip/internal/application/commands"
	"github.com/d6o/homeclip/internal/application/dtos"
	"github.com/d6o/homeclip/internal/application/queries"
)

type DocumentApplicationService struct {
	updateContentHandler application.UpdateContentCommandHandler
	getContentHandler    application.GetContentQueryHandler
}

func NewDocumentApplicationService(
	updateContentHandler application.UpdateContentCommandHandler,
	getContentHandler application.GetContentQueryHandler,
) *DocumentApplicationService {
	return &DocumentApplicationService{
		updateContentHandler: updateContentHandler,
		getContentHandler:    getContentHandler,
	}
}

func (s *DocumentApplicationService) GetContent(ctx context.Context, documentID string) (*dtos.GetContentResponse, error) {
	query := queries.GetContentQuery{
		DocumentID: documentID,
	}
	return s.getContentHandler.Handle(ctx, query)
}

func (s *DocumentApplicationService) UpdateContent(ctx context.Context, documentID string, content string) (*dtos.UpdateContentResponse, error) {
	cmd := commands.UpdateContentCommand{
		DocumentID: documentID,
		Content:    content,
	}
	return s.updateContentHandler.Handle(ctx, cmd)
}