package container

import (
	"embed"

	"github.com/d6o/homeclip/internal/application/commands"
	"github.com/d6o/homeclip/internal/application/queries"
	"github.com/d6o/homeclip/internal/application/services"
	"github.com/d6o/homeclip/internal/domain/repositories"
	domainservices "github.com/d6o/homeclip/internal/domain/services"
	"github.com/d6o/homeclip/internal/infrastructure/config"
	"github.com/d6o/homeclip/internal/infrastructure/http/handlers"
	"github.com/d6o/homeclip/internal/infrastructure/http/router"
	"github.com/d6o/homeclip/internal/infrastructure/http/server"
	"github.com/d6o/homeclip/internal/infrastructure/persistence"
	infraservices "github.com/d6o/homeclip/internal/infrastructure/services"
)

type Container struct {
	Config                *config.Config
	DocumentRepository    repositories.DocumentRepository
	FileStorageRepository repositories.FileStorageRepository
	DocumentService       *domainservices.DocumentService
	ExpirationService     *domainservices.ExpirationService
	CleanupService        *infraservices.CleanupService
	UpdateContentHandler  *commands.UpdateContentCommandHandler
	GetContentHandler     *queries.GetContentQueryHandler
	UploadFileHandler     *commands.UploadFileCommandHandler
	DeleteFileHandler     *commands.DeleteFileCommandHandler
	GetFileHandler        *queries.GetFileQueryHandler
	ListFilesHandler      *queries.ListFilesQueryHandler
	DocumentAppService    *services.DocumentApplicationService
	DocumentHandler       *handlers.DocumentHandler
	FileHandler           *handlers.FileHandler
	Router                *router.Router
	Server                *server.Server
}

func NewContainer(staticFiles embed.FS) *Container {
	c := &Container{}

	c.Config = config.LoadConfig()

	c.DocumentRepository = persistence.NewMemoryDocumentRepository()
	c.FileStorageRepository = persistence.NewMemoryFileStorage()

	c.DocumentService = domainservices.NewDocumentService(c.DocumentRepository)
	c.ExpirationService = domainservices.NewExpirationService(c.DocumentRepository, c.FileStorageRepository)

	c.CleanupService = infraservices.NewCleanupService(c.DocumentRepository, c.FileStorageRepository, c.ExpirationService, c.Config.CleanupInterval)

	c.UpdateContentHandler = commands.NewUpdateContentCommandHandler(c.DocumentService)
	c.GetContentHandler = queries.NewGetContentQueryHandler(c.DocumentService)
	c.UploadFileHandler = commands.NewUploadFileCommandHandler(c.DocumentService, c.DocumentRepository, c.FileStorageRepository)
	c.DeleteFileHandler = commands.NewDeleteFileCommandHandler(c.DocumentRepository, c.FileStorageRepository)
	c.GetFileHandler = queries.NewGetFileQueryHandler(c.DocumentRepository, c.FileStorageRepository)
	c.ListFilesHandler = queries.NewListFilesQueryHandler(c.DocumentRepository)

	c.DocumentAppService = services.NewDocumentApplicationService(
		c.UpdateContentHandler,
		c.GetContentHandler,
	)

	c.DocumentHandler = handlers.NewDocumentHandler(c.DocumentAppService)
	c.FileHandler = handlers.NewFileHandler(
		c.UploadFileHandler,
		c.DeleteFileHandler,
		c.GetFileHandler,
		c.ListFilesHandler,
	)
	c.Router = router.NewRouter(c.DocumentHandler, c.FileHandler, staticFiles, c.Config.EnableFileUploads)

	handler := c.Router.Setup()
	c.Server = server.NewServer(handler, c.Config)

	return c
}

func (c *Container) StartServer() error {
	return c.Server.Start()
}
