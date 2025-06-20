package urlmanager

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/limbo/url_shortener/internal/api"
)

type PgxPool interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Generator is used to generate short codes for links
type Generator interface {
	Gen() string
}

// Main class for working with PostgreSQL
type Client struct {
	pool          PgxPool
	codeGenerator Generator
}

type DBCfg struct {
	Address  string
	User     string
	Password string
	DBName   string
	Options  map[string]string
}

// Returns new DB client with provided short code generator
func New(cfg DBCfg, gen Generator) *Client {
	optsStr := ""
	if len(cfg.Options) != 0 {
		optsStr = "?"
		for k, v := range cfg.Options {
			optsStr += k + "=" + v
		}
	}
	p, err := pgxpool.New(context.Background(), "postgresql://"+cfg.User+":"+cfg.Password+"@"+cfg.Address+"/"+cfg.DBName+optsStr)
	if err != nil {
		log.Fatal(err)
	}
	err = p.Ping(context.Background())
	if err != nil {
		log.Fatal("ping error: " + err.Error())
	}
	return &Client{
		pool:          p,
		codeGenerator: gen,
	}
}

func NewWithPool(p PgxPool, gen Generator) *Client {
	return &Client{
		pool:          p,
		codeGenerator: gen,
	}
}

// Generates new shortCode for provided link, saves it to DB
// and returns it
func (c *Client) SaveURL(link string) (string, error) {
	var shortCode string
	tx, err := c.pool.Begin(context.Background())
	if err != nil {
		return "", errors.New("tx error: " + err.Error())
	}
	defer tx.Rollback(context.Background())

	err = tx.QueryRow(context.Background(), `SELECT short_code FROM redirects WHERE link = $1;`, link).Scan(&shortCode)
	if err != nil && err != pgx.ErrNoRows {
		return "", errors.New("error looking up for a code: " + err.Error())
	} else if err == pgx.ErrNoRows {
		shortCode = c.codeGenerator.Gen()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_, err = tx.Exec(ctx, `INSERT INTO redirects (link, short_code) VALUES ($1, $2);`, link, shortCode)
		if err != nil {
			if isDuplicateFieldError(err) {
				return "", errors.New("duplicating short code")
			}
			return "", errors.New("error inserting code: " + err.Error())
		}
	}
	err = tx.Commit(context.Background())
	if err != nil {
		return "", errors.New("error tx commit: " + err.Error())
	}
	return shortCode, nil
}

// Searchs link with provided shortcode, returns it
func (c *Client) GetLinkByCode(code string) (string, error) {
	var link string
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	row := c.pool.QueryRow(ctx, `SELECT link FROM redirects WHERE short_code = $1;`, code)
	if err := row.Scan(&link); err != nil {
		if err == pgx.ErrNoRows {
			return "", api.ErrNoRow
		}
		return "", errors.New("error getting link: " + err.Error())
	}
	return link, nil
}
