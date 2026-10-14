package app

import (
	"context"
	"fmt"
	"main/config"
	"main/internal/contoller/http/handlers"
	"main/internal/repository"
	"main/internal/service"
	"main/pkg/postgres"
)

type (
	Container struct {
		Controllers  *Controllers
		Servicies    *Services
		Repositories *Repositories
		Drivers      *Drivers
	}

	Controllers struct {
		Users     *handlers.UserHandler
		Responses *handlers.ResponseHandler
	}

	Services struct {
		Users     *service.UserService
		Responses *service.ResponseService
	}

	Repositories struct {
		Users     *repository.UserRepository
		Responses *repository.ResponseRepository
	}

	Drivers struct {
		Postgres *postgres.Client
	}
)

func (c *Container) Close() {
	c.Drivers.Postgres.Close()
}

func NewContainer(ctx context.Context, cfg *config.Config) (*Container, error) {
	drivers, err := initDrivers(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("initDrivers: %w", err)
	}

	repositories := initRepositories(drivers)
	services := initServices(repositories)
	controllers := initControllers(services)

	return &Container{
		Controllers:  controllers,
		Servicies:    services,
		Repositories: repositories,
		Drivers:      drivers,
	}, nil
}

func initControllers(services *Services) *Controllers {
	return &Controllers{
		Users:     handlers.NewUserHandler(services.Users),
		Responses: handlers.NewResponseHandler(services.Responses),
	}
}

func initServices(repositories *Repositories) *Services {
	return &Services{
		Users:     service.NewUserService(repositories.Users),
		Responses: service.NewResponseService(repositories.Responses, repositories.Users),
	}
}

func initRepositories(drivers *Drivers) *Repositories {
	return &Repositories{
		Users:     repository.NewUserRepository(drivers.Postgres),
		Responses: repository.NewResponseRepository(drivers.Postgres),
	}
}

func initDrivers(ctx context.Context, cfg *config.Config) (*Drivers, error) {
	dsn := createDsn(cfg)
	driver, err := postgres.NewClient(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres.NewClient: %w", err)
	}

	return &Drivers{
		Postgres: driver,
	}, nil
}

func createDsn(cfg *config.Config) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DatabaseSecret.Host,
		cfg.DatabaseSecret.Port,
		cfg.DatabaseSecret.User,
		cfg.DatabaseSecret.Password,
		cfg.DatabaseSecret.Name,
	)
}
