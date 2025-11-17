package testcontainer

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/docker/docker/api/types/container"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const sqlExt = ".sql"

// nolint:nestif
func CreatePostgresStorage(ctx context.Context, logger logrus.FieldLogger) (*pgxpool.Pool, *postgres.PostgresContainer, error) {
	const (
		defaultPostgres = "postgres"
		postgresImage   = "romashqua/romashquarepo:postgres-16-conf"
	)

	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, nil, errors.Wrap(err, "findProjectRoot")
	}

	migrationsPath := filepath.Join(projectRoot, "migrations")

	migrationFiles, err := getMigrationFiles(migrationsPath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "getMigrationFiles")
	}

	postgresContainer, err := postgres.Run(ctx,
		postgresImage,
		postgres.WithInitScripts(migrationFiles...),
		postgres.BasicWaitStrategies(),
		testcontainers.WithTmpfs(map[string]string{"/var/lib/postgresql/data": ""}),
		testcontainers.WithHostConfigModifier(func(hostConfig *container.HostConfig) {
			hostConfig.ShmSize = 1024 * 1024 * 256
		}),
	)
	if err != nil {
		if postgresContainer != nil {
			logs, logErr := postgresContainer.Logs(ctx)
			if logErr != nil {
				logger.Errorf("postgresContainer.Logs: %v", logErr)
			} else {
				logsBytes, readErr := io.ReadAll(logs)
				if readErr != nil {
					logger.Errorf("io.ReadAll: %v", readErr)
				} else {
					logger.Infof("postgres logs: %v", string(logsBytes))
				}
			}

			if terminateErr := postgresContainer.Terminate(ctx); terminateErr != nil {
				logger.Errorf("failed to terminate postgres container: %v", terminateErr)
			}
		}

		return nil, nil, errors.Wrap(err, "postgres.Run")
	}

	if postgresContainer == nil {
		return nil, nil, errors.New("postgres container is nil")
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "postgresContainer.Host")
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, nil, errors.Wrap(err, "postgresContainer.MappedPort")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		defaultPostgres, defaultPostgres, host, port.Port(), defaultPostgres)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, errors.Wrap(err, "pgxpool.ParseConfig")
	}

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "pgxpool.NewWithConfig")
	}

	return db, postgresContainer, nil
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "os.Getwd")
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}

		dir = parentDir
	}

	return "", errors.Wrap(errors.New("root not found - go.mod not found)"), "findProjectRoot")
}

func getMigrationFiles(migrationsDir string) ([]string, error) {
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, errors.Wrap(err, "os.ReadDir")
	}

	var migrationFiles []string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if filepath.Ext(filename) == sqlExt {
			migrationFiles = append(migrationFiles, filepath.Join(migrationsDir, filename))
		}
	}

	sort.Strings(migrationFiles)

	return migrationFiles, nil
}
