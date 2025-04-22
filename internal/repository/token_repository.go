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
		ID:         omit.FromPtr(params.ID),
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

func DeleteToken(ctx context.Context, db bob.Executor, token string) error {
	_, err := models.Tokens.Delete(
		psql.WhereAnd(
			models.DeleteWhere.Tokens.Token.EQ(token),
			models.DeleteWhere.Tokens.Expires.GT(time.Now()),
		),
	).Exec(ctx, db)
	return err
}
func GetToken(ctx context.Context, db bob.Executor, token string) (*models.Token, error) {
	res, err := models.Tokens.Query(
		models.SelectWhere.Tokens.Token.EQ(token),
	).One(ctx, db)
	if err != nil {
		return nil, err
	}
	return res, nil
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
