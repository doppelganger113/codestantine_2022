package test

import (
	"api"
	"api/logger"
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"testing"
	"time"
)

func closeAfterTime(err chan<- error) {
	time.Sleep(10 * time.Second)
	err <- fmt.Errorf("closed")
}

func TestIntegrationAppGetImages(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5434/tcp", "5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "example",
			"POSTGRES_DB":       "db",
		},
	}
	postgreContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	postgreEndpoint, err := postgreContainer.Endpoint(context.Background(), "")
	if err != nil {
		t.Error(err)
	}

	defer func(postgreContainer testcontainers.Container, ctx context.Context) {
		terminationErr := postgreContainer.Terminate(ctx)
		if terminationErr != nil {
			t.Error(terminationErr)
		}
	}(postgreContainer, ctx)

	log.Println("Bootstrapping...")

	logger := logger.NewLogger(logger.WithPretty())

	// Application
	app, err := api.InitializeAppForTesting(logger)
	if err != nil {
		t.Fatalf("failed initializing app: %s", err)
	}
	app.Config.DatabaseUrl = fmt.Sprintf("postgresql://postgres:example@%s/db", postgreEndpoint)
	api.Bootstrap(app, logger, api.WithInterrupt(TimeoutInterrupt(10*time.Second)))

	log.Println("Ended")
}
