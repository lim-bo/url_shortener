package urlmanager_test

import (
	"context"
	"regexp"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	urlmanager "github.com/limbo/url_shortener/internal/url_manager"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// setting up test postgresql container
func setupTestDB(t *testing.T) (urlmanager.DBCfg, func()) {
	container, err := postgres.Run(context.Background(), "postgres:17",
		postgres.WithUsername("test_user"),
		postgres.WithDatabase("url_shortener"),
		postgres.WithPassword("test_password"))
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
	return urlmanager.DBCfg{
			Address:  pgxpoolCfg.ConnConfig.Host + ":" + strconv.FormatUint(uint64(pgxpoolCfg.ConnConfig.Port), 10),
			Password: "test_password",
			User:     "test_user",
			DBName:   "url_shortener",
		}, func() {
			container.Terminate(context.Background())
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
	mockPool.ExpectExec(`INSERT INTO`).WithArgs(testUrl, testCode).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	shortCode, err := cli.SaveURL(testUrl)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, shortCode, testCode)
}

func TestGetLink(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPool.Close()

	cli := urlmanager.NewWithPool(mockPool, &CodeGenMock{})
	expectedSQL := regexp.QuoteMeta("SELECT link FROM redirects WHERE short_code = $1;")
	mockPool.ExpectQuery(expectedSQL).
		WithArgs(testCode).
		WillReturnRows(pgxmock.NewRows([]string{"link"}).AddRow(testUrl))
	link, err := cli.GetLinkByCode(testCode)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, link, testUrl)
}
