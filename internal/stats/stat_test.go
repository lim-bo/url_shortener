package stats_test

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/limbo/url_shortener/internal/stats"
	"github.com/stretchr/testify/assert"
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
	t.Run("successful with no insertion", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`ALTER TABLE`).WithArgs(code).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := cli.IncreaseClicks(link, code)
		assert.NoError(t, err)
	})
	t.Run("successful with insertion", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`ALTER TABLE`).WithArgs(code).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(`INSERT INTO`).WithArgs(link, code).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := cli.IncreaseClicks(link, code)
		assert.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`ALTER TABLE`).WithArgs(code).WillReturnError(errors.New("test error"))
		mock.ExpectRollback()
		err := cli.IncreaseClicks(link, code)
		assert.Error(t, err)
	})
}

func TestGetStat(t *testing.T) {

}

func TestIntegrational(t *testing.T) {

}
