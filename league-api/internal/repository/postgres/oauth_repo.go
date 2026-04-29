package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	idb "league-api/internal/db"
	"league-api/internal/model"
	"league-api/internal/repository"
)

type oauthRepo struct {
	pool *sqlx.DB
}

func NewOAuthRepo(db *sqlx.DB) repository.OAuthAccountRepository {
	return &oauthRepo{pool: db}
}

func (r *oauthRepo) db(ctx context.Context) idb.DBTX {
	if tx := idb.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *oauthRepo) GetByProviderSub(ctx context.Context, provider, sub string) (*model.OAuthAccount, error) {
	var a model.OAuthAccount
	err := r.db(ctx).GetContext(ctx, &a,
		`SELECT * FROM user_oauth_accounts WHERE provider = $1 AND provider_sub = $2`,
		provider, sub,
	)
	if err != nil {
		return nil, fmt.Errorf("oauthRepo.GetByProviderSub: %w", err)
	}
	return &a, nil
}

func (r *oauthRepo) Create(ctx context.Context, a *model.OAuthAccount) error {
	const q = `
		INSERT INTO user_oauth_accounts (user_id, provider, provider_sub)
		VALUES ($1, $2, $3)`
	_, err := r.db(ctx).ExecContext(ctx, q, a.UserID, a.Provider, a.ProviderSub)
	if err != nil {
		return fmt.Errorf("oauthRepo.Create: %w", err)
	}
	return nil
}

func (r *oauthRepo) ListByUser(ctx context.Context, userID int64) ([]model.OAuthAccount, error) {
	var accounts []model.OAuthAccount
	err := r.db(ctx).SelectContext(ctx, &accounts,
		`SELECT * FROM user_oauth_accounts WHERE user_id = $1`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("oauthRepo.ListByUser: %w", err)
	}
	return accounts, nil
}
