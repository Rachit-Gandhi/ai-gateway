// Package testutil provides shared test helpers.
package testutil

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// PostgresContainer holds the running container and its DSN.
type PostgresContainer struct {
	DSN       string
	container *postgres.PostgresContainer
}

// StartPostgres starts a Postgres 16 container and returns its DSN.
// Returns an error if Docker is not available or the container fails to start.
func StartPostgres() (pg *PostgresContainer, err error) {
	if _, lookErr := exec.LookPath("docker"); lookErr != nil {
		return nil, fmt.Errorf("docker not found in PATH: %w", lookErr)
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("testcontainers panic: %v", r)
		}
	}()

	ctx := context.Background()

	container, startErr := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("ferrogw_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	if startErr != nil {
		return nil, fmt.Errorf("start postgres container: %w", startErr)
	}

	dsn, connErr := container.ConnectionString(ctx, "sslmode=disable")
	if connErr != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get connection string: %w", connErr)
	}

	return &PostgresContainer{DSN: dsn, container: container}, nil
}

// Terminate stops and removes the container.
func (c *PostgresContainer) Terminate() {
	if c != nil && c.container != nil {
		_ = c.container.Terminate(context.Background())
	}
}
