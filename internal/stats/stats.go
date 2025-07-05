package stats

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/limbo/url_shortener/internal/api"
	"github.com/limbo/url_shortener/models"
)

type ClicksDBCli struct {
	conn   *sql.DB
	dbName string
}

type ClicksDBCfg struct {
	Address  string
	Username string
	Password string
	Database string
}

func New(cfg ClicksDBCfg) *ClicksDBCli {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{cfg.Address},
		Auth: clickhouse.Auth{
			Username: cfg.Username,
			Password: cfg.Password,
			Database: cfg.Database,
		},
		DialTimeout: time.Second * 5,
	})
	if err := conn.Ping(); err != nil {
		log.Fatal(err)
	}
	conn.SetMaxOpenConns(15)
	conn.SetMaxIdleConns(10)
	return &ClicksDBCli{
		conn:   conn,
		dbName: cfg.Database,
	}
}

func NewWithConn(conn *sql.DB, dbName string) *ClicksDBCli {
	return &ClicksDBCli{
		conn:   conn,
		dbName: dbName,
	}
}

func (cli *ClicksDBCli) IncreaseClicks(link, code string) error {
	_, err := cli.conn.Exec(`INSERT INTO `+cli.dbName+`.redirect_stat (link, code, clicks) 
VALUES (?, ?, 1);`, link, code)
	if err != nil {
		return errors.New("updating clicks error: " + err.Error())
	}
	return nil
}

func (cli *ClicksDBCli) GetStats(code string) (*models.ClicksStat, error) {
	var stat models.ClicksStat
	row := cli.conn.QueryRow(`SELECT link, code, sum(clicks) FROM `+cli.dbName+`.redirect_stat WHERE code = ? GROUP BY link, code;`, code)
	if err := row.Scan(&stat.OGLink, &stat.Code, &stat.Clicks); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, api.ErrNoRedirects
		}
		return nil, errors.New("error getting stat values: " + err.Error())
	}
	return &stat, nil
}
