package token

import (
	"context"
)

type AccessTokenMinter interface {
	MintAccessToken(ctx context.Context, userID string) (string, error)
}
