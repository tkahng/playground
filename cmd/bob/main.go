package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
)

func main() {
	ctx := context.Background()
	// conf := conf.AppConfigGetter()
	userId := uuid.New()
	params := &repository.TokenDTO{
		UserID:     &userId,
		Type:       models.TokenTypesVerificationToken,
		Identifier: "tkahng@gmail.com",
		Expires:    time.Now().Add(time.Hour * 24),
		Token:      uuid.NewString(),
	}

	// dbx := core.NewBobFromConf(ctx, conf.Db)
	q := models.Tokens.Query()
	// q.Apply(
	// 	models.SelectWhere.Tokens.Type.EQ(params.Type),
	// 	models.SelectWhere.Tokens.Identifier.EQ(params.Identifier),
	// 	models.SelectWhere.Tokens.UserID.EQ(params.UserID),
	// )
	q.Apply(
		models.SelectWhere.Tokens.Type.EQ(params.Type),
		models.SelectWhere.Tokens.Identifier.EQ(params.Identifier),
		// models.SelectWhere.Tokens.UserID.EQ(params.UserID),
		// sm.OrderBy(models.TokenColumns.)
	)
	if params.UserID != nil {
		q.Apply(
			models.SelectWhere.Tokens.UserID.EQ(*params.UserID),
		)
	} else {
		q.Apply(
			models.SelectWhere.Tokens.UserID.IsNull(),
		)
	}
	q.Apply(
		sm.OrderBy(models.TokenColumns.CreatedAt),
		sm.OrderBy(models.TokenColumns.ID),
	)
	queryString, args, err := q.Build(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(queryString, args)
	// q.WriteQuery()

	// data, err := q.All(ctx, dbx)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(len(data))
	// for _, user := range data {
	// 	fmt.Println(len(user))
	// }
}
