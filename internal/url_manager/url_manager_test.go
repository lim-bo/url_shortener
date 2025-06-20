package urlmanager_test

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	urlmanager "github.com/limbo/url_shortener/internal/url_manager"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setting up test postgresql container
func setupTestDB(t *testing.T) urlmanager.DBCfg {
	container, err := postgres.Run(context.Background(), "postgres:17",
		postgres.WithUsername("test_user"),
		postgres.WithDatabase("url_shortener"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatal("error running test container: " + err.Error())
	}
	connStr, err := container.ConnectionString(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	pgxpoolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		t.Fatal("error parsing config: " + err.Error())
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxpoolCfg)
	if err != nil {
		t.Fatal("error connecting to container: " + err.Error())
	}
	_, err = pool.Exec(context.Background(), `CREATE TABLE "redirects" (
id serial primary key,
link TEXT NOT NULL,
short_code VARCHAR(8) NOT NULL UNIQUE);`)
	if err != nil {
		t.Fatal("error setting migrations: " + err.Error())
	}
	pool.Close()
	t.Cleanup(func() {
		container.Terminate(context.Background())
	})
	return urlmanager.DBCfg{
		Address:  pgxpoolCfg.ConnConfig.Host + ":" + strconv.FormatUint(uint64(pgxpoolCfg.ConnConfig.Port), 10),
		Password: "test_password",
		User:     "test_user",
		DBName:   "url_shortener",
	}
}

var (
	testCode = "abcd1234"
	testUrl  = "https://google.com"
)

type CodeGenMock struct {
}

func (gen *CodeGenMock) Gen() string {
	return testCode
}
func TestSaveURL(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPool.Close()

	cli := urlmanager.NewWithPool(mockPool, &CodeGenMock{})
	t.Run("code with no inserting", func(t *testing.T) {
		mockPool.ExpectBegin()

		mockPool.ExpectQuery(`SELECT short_code`).WithArgs(testUrl).WillReturnRows(pgxmock.NewRows([]string{"short_code"}).AddRow(testCode))

		mockPool.ExpectCommit()

		shortCode, err := cli.SaveURL(testUrl)
		if err != nil {
			t.Error(err)
		}
		assert.NoError(t, mockPool.ExpectationsWereMet())
		assert.Equal(t, shortCode, testCode)
	})
	t.Run("code with inserting", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.ExpectQuery(`SELECT short_code`).WithArgs(testUrl).WillReturnError(pgx.ErrNoRows)
		mockPool.ExpectExec(`INSERT INTO`).WithArgs(testUrl, testCode).WillReturnResult(pgxmock.NewResult("INSERT", 1))
		mockPool.ExpectCommit()
		shortCode, err := cli.SaveURL(testUrl)
		if err != nil {
			t.Error(err)
		}
		assert.NoError(t, mockPool.ExpectationsWereMet())
		assert.Equal(t, shortCode, testCode)
	})
	t.Run("mock error", func(t *testing.T) {
		mockPool.ExpectExec(`INSERT INTO`).WithArgs(testUrl, testCode).WillReturnError(errors.New("repository error"))
		_, err := cli.SaveURL(testUrl)
		assert.Error(t, err)
	})
}

func TestGetLink(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPool.Close()

	cli := urlmanager.NewWithPool(mockPool, &CodeGenMock{})
	expectedSQL := regexp.QuoteMeta("SELECT link FROM redirects WHERE short_code = $1;")
	t.Run("successful link recieving", func(t *testing.T) {
		mockPool.ExpectQuery(expectedSQL).
			WithArgs(testCode).
			WillReturnRows(pgxmock.NewRows([]string{"link"}).AddRow(testUrl))
		link, err := cli.GetLinkByCode(testCode)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, link, testUrl)
	})
	t.Run("getting link error", func(t *testing.T) {
		mockPool.ExpectQuery(expectedSQL).
			WithArgs(testCode).
			WillReturnError(errors.New("repository error"))
		_, err := cli.GetLinkByCode(testCode)
		assert.Error(t, err)
	})
}

func TestIntegrational(t *testing.T) {
	cfg := setupTestDB(t)
	cli := urlmanager.New(cfg, &urlmanager.CodeGenerator{})

	code, err := cli.SaveURL(testUrl)
	if err != nil {
		t.Fatal(err)
	}

	originalLink, err := cli.GetLinkByCode(code)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, testUrl, originalLink)
}
