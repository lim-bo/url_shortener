package urlmanager

import "github.com/jackc/pgconn"

func isDuplicateFieldError(err error) bool {
	pgErr, ok := err.(*pgconn.PgError)
	return ok && pgErr.Code == "23505"
}
