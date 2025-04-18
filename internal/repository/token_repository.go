package repository

import (
	"context"
	"errors"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/tkahng/authgo/internal/db/models"
)

type TokenDTO struct {
	Type       models.TokenTypes `db:"type" json:"type"`
	Identifier string            `db:"identifier" json:"identifier"`
	Expires    time.Time         `db:"expires" json:"expires"`
	Token      string            `db:"token" json:"token"`
	ID         *uuid.UUID        `db:"id" json:"id"`
	UserID     *uuid.UUID        `db:"user_id" json:"user_id"`
	Otp        *string           `db:"otp" json:"otp"`
}

type OtpDto struct {
	Type       models.TokenTypes `db:"type" json:"type"`
	Identifier string            `db:"identifier" json:"identifier"`
	Otp        *string           `db:"otp" json:"otp"`
	UserID     *uuid.UUID        `db:"user_id" json:"user_id"`
}

func CreateToken(ctx context.Context, db bob.Executor, params *TokenDTO) (*models.Token, error) {
	if params == nil {
		return nil, errors.New("params is nil")
	}
	newVar := &models.TokenSetter{
		UserID:     omitnull.FromPtr(params.UserID),
		Type:       omit.From(params.Type),
		Identifier: omit.From(params.Identifier),
		Expires:    omit.From(params.Expires),
		Token:      omit.From(params.Token),
		Otp:        omitnull.FromPtr(params.Otp),
	}
	if params.ID != nil {
		newVar.ID = omit.FromPtr(params.ID)
	}
	if params.UserID != nil {
		newVar.UserID = omitnull.FromPtr(params.UserID)

	}
	return models.Tokens.Insert(newVar, im.Returning("*")).One(ctx, db)
}

func UseToken(ctx context.Context, db bob.Executor, params string) (*models.Token, error) {
	token, err := models.
		Tokens.
		Delete(
			psql.WhereAnd(
				models.DeleteWhere.Tokens.Token.EQ(params),
				models.DeleteWhere.Tokens.Expires.GT(time.Now()),
			),
			dm.Returning("*"),
		).
		One(ctx, db)
	return token, err
}

func FindFirstTokenByUserAndType(ctx context.Context, db bob.Executor, params *OtpDto) (*models.Token, error) {
	if params == nil {
		return nil, errors.New("params is nil")
	}
	q := models.Tokens.Query()
	q.Apply(
		models.SelectWhere.Tokens.Type.EQ(params.Type),
		models.SelectWhere.Tokens.Identifier.EQ(params.Identifier),
		models.SelectWhere.Tokens.Expires.GT(time.Now()),
		sm.OrderBy(models.TokenColumns.CreatedAt).Desc(),
		sm.OrderBy(models.TokenColumns.ID).Desc(),
		sm.Limit(1),
	)
	return q.One(ctx, db)
}

func DeleteTokensByUser(ctx context.Context, db bob.Executor, params *OtpDto) error {
	if params == nil {
		return errors.New("params is nil")
	}
	q := models.
		Tokens.
		Delete()
	q.Apply(
		models.DeleteWhere.Tokens.Identifier.EQ(params.Identifier),
		models.DeleteWhere.Tokens.Type.EQ(params.Type),
	)
	_, err := q.Exec(ctx, db)

	return err
}
