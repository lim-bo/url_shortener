package stats

import (
	"database/sql"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClicksDBCli struct {
	txmu   sync.Mutex
	conn   *sql.DB
	dbName string
}

type ClicksDBCfg struct {
	Address  string
	Username string
	Password string
	Database string
}

type ClicksStat struct {
	Code   string
	OGLink string
	Clicks uint64
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
		txmu:   sync.Mutex{},
		dbName: cfg.Database,
	}
}

func NewWithConn(conn *sql.DB, dbName string) *ClicksDBCli {
	return &ClicksDBCli{
		conn:   conn,
		txmu:   sync.Mutex{},
		dbName: dbName,
	}
}

func (cli *ClicksDBCli) IncreaseClicks(link, code string) error {
	cli.txmu.Lock()
	defer cli.txmu.Unlock()
	tx, err := cli.conn.Begin()
	if err != nil {
		return errors.New("tx error: " + err.Error())
	}
	defer tx.Rollback()
	result, err := tx.Exec(`ALTER TABLE `+cli.dbName+`.redirect_stat UPDATE clicks += 1 WHERE code = ?;`, code)
	if err != nil {
		return errors.New("updating clicks error: " + err.Error())
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		_, err = tx.Exec(`INSERT INTO `+cli.dbName+`.redirect_stat (link, code, clicks) VALUES (?, ?, 1);`, link, code)
		if err != nil {
			return errors.New("inserting stat error: " + err.Error())
		}
	}
	err = tx.Commit()
	if err != nil {
		return errors.New("commiting error: " + err.Error())
	}
	return nil
}

func (cli *ClicksDBCli) GetStats(link, code string) (*ClicksStat, error) {
	var stat ClicksStat
	row := cli.conn.QueryRow(`SELECT link, code, clicks FROM `+cli.dbName+`.redirect_stat WHERE code = ?;`, code)
	if err := row.Scan(&stat.OGLink, &stat.Code, &stat.Clicks); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &ClicksStat{
				OGLink: link,
				Code:   code,
				Clicks: 0,
			}, nil
		}
		return nil, errors.New("error getting stat values: " + err.Error())
	}
	return &stat, nil
}
