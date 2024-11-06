// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: session.sql

package sqlc

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createSession = `-- name: CreateSession :one
INSERT INTO sessions(
    id,
    username,
 refresh_token,
 user_agent,
 client_ip,
 is_blocked,
 expired_at
)
VALUES(
    $1,$2,$3,$4,$5,$6,$7
) RETURNING id, username, refresh_token, user_agent, client_ip, is_blocked, expired_at, create_at
`

type CreateSessionParams struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiredAt    time.Time `json:"expired_at"`
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error) {
	row := q.db.QueryRow(ctx, createSession,
		arg.ID,
		arg.Username,
		arg.RefreshToken,
		arg.UserAgent,
		arg.ClientIp,
		arg.IsBlocked,
		arg.ExpiredAt,
	)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.RefreshToken,
		&i.UserAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.ExpiredAt,
		&i.CreateAt,
	)
	return i, err
}

const getSession = `-- name: GetSession :one
SELECT id, username, refresh_token, user_agent, client_ip, is_blocked, expired_at, create_at FROM sessions
WHERE id=$1 LIMIT 1
`

func (q *Queries) GetSession(ctx context.Context, id uuid.UUID) (Session, error) {
	row := q.db.QueryRow(ctx, getSession, id)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.RefreshToken,
		&i.UserAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.ExpiredAt,
		&i.CreateAt,
	)
	return i, err
}
