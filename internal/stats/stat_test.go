package stats_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/limbo/url_shortener/internal/stats"
	"github.com/limbo/url_shortener/models"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	link   = "google.com"
	code   = "abcd1234"
	clicks = 15
)

func TestUpdateStat(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	cli := stats.NewWithConn(db, "testing")
	expectedQuery := regexp.QuoteMeta(`INSERT INTO testing.redirect_stat (link, code, clicks) 
VALUES (?, ?, 1);`)
	t.Run("successful", func(t *testing.T) {
		mock.ExpectExec(expectedQuery).WithArgs(link, code).WillReturnResult(sqlmock.NewResult(1, 1))
		err := cli.IncreaseClicks(link, code)
		assert.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		mock.ExpectExec(expectedQuery).WithArgs(link, code).WillReturnError(errors.New("test error"))
		err := cli.IncreaseClicks(link, code)
		assert.Error(t, err)
	})
}

func TestGetStat(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	cli := stats.NewWithConn(db, "testing")
	t.Run("successful", func(t *testing.T) {
		expectedSQL := regexp.QuoteMeta(`SELECT link, code, sum(clicks) FROM testing.redirect_stat WHERE code = ? GROUP BY link, code;`)
		mock.ExpectQuery(expectedSQL).
			WithArgs(code).
			WillReturnRows(sqlmock.NewRows([]string{"link", "code", "clicks"}).AddRow(link, code, clicks))
		stat, err := cli.GetStats(link, code)
		assert.NoError(t, err)
		assert.Equal(t, &models.ClicksStat{
			Code:   code,
			OGLink: link,
			Clicks: uint64(clicks),
		}, stat)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT link, code, clicks`).
			WithArgs(code).
			WillReturnError(errors.New("test error"))
		_, err := cli.GetStats(link, code)
		assert.Error(t, err)
	})
}

func TestIntegrational(t *testing.T) {
	conn := setupContainer(t)
	cli := stats.NewWithConn(conn, "default")
	err := cli.IncreaseClicks(link, code)
	assert.NoError(t, err)
	stat, err := cli.GetStats(link, code)
	assert.NoError(t, err)
	assert.Equal(t, models.ClicksStat{
		OGLink: link,
		Code:   code,
		Clicks: 1,
	}, *stat)
}

func setupContainer(t *testing.T) *sql.DB {
	req := testcontainers.ContainerRequest{
		Image:        "clickhouse/clickhouse-server:23.8",
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"CLICKHOUSE_DB":       "default",
			"CLICKHOUSE_USER":     "default",
			"CLICKHOUSE_PASSWORD": "password",
		},
		WaitingFor: wait.ForListeningPort("9000/tcp").WithStartupTimeout(30 * time.Second),
	}
	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	host, err := container.Host(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	port, err := container.MappedPort(context.Background(), "9000")
	if err != nil {
		t.Fatal(err)
	}
	db := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{host + ":" + port.Port()},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "password",
		},
	})
	for i := 0; i < 10; i++ {
		err := db.Ping()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS default.redirect_stat (
    link String, 
    code String,
    clicks UInt64
) ENGINE = SummingMergeTree() ORDER BY clicks;`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		container.Terminate(context.Background())
	})
	return db
}
