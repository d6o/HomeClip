package services

import (
	"context"
	"errors"

	"github.com/d6o/homeclip/internal/domain/entities"
	"github.com/d6o/homeclip/internal/domain/repositories"
	"github.com/d6o/homeclip/internal/domain/valueobjects"
)

type DocumentService struct {
	repository repositories.DocumentRepository
}

func NewDocumentService(repository repositories.DocumentRepository) *DocumentService {
	return &DocumentService{
		repository: repository,
	}
}

func (s *DocumentService) GetOrCreateDocument(ctx context.Context, id entities.DocumentID) (*entities.Document, error) {
	document, err := s.repository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, entities.ErrDocumentNotFound) {
			document = entities.NewDocument(id)
			if err := s.repository.Save(ctx, document); err != nil {
				return nil, err
			}
			return document, nil
		}
		return nil, err
	}
	return document, nil
}

func (s *DocumentService) UpdateDocumentContent(ctx context.Context, id entities.DocumentID, contentValue string) (*entities.Document, error) {
	content, err := valueobjects.NewContent(contentValue)
	if err != nil {
		return nil, err
	}

	document, err := s.GetOrCreateDocument(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := document.UpdateContent(content); err != nil {
		return nil, err
	}

	if err := s.repository.Save(ctx, document); err != nil {
		return nil, err
	}

	return document, nil
}
