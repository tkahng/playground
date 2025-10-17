package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tkahng/playground/internal/database/repository"
)

type A struct {
	_   struct{} `db:"a" json:"-"`
	ID  string   `db:"id,pk" json:"id"`
	Age int64    `db:"age" json:"age"`
	B   *B       `db:"b" src:"id" dest:"a_id" table:"public.b" json:"b,omitempty"`
	Cs  []*C     `db:"cs" src:"id" dest:"a_id" table:"public.c" json:"cs,omitempty"`
	Ds  []*D     `db:"ds" src:"id" dest:"id" table:"public.d" through:"public.ad" through_src:"a_id" through_dest:"d_id" json:"ds,omitempty"`
}
type B struct {
	_    struct{} `db:"b" json:"-"`
	ID   string   `db:"id" json:"id"`
	Body string   `db:"body" json:"body"`
	AID  string   `db:"a_id" json:"a_id"`
	A    *A       `db:"a" src:"a_id" dest:"id" table:"public.a" json:"a,omitempty"`
}
type C struct {
	_    struct{} `db:"c" json:"-"`
	ID   string   `db:"id" json:"id"`
	Code int64    `db:"code" json:"code"`
	AID  string   `db:"a_id" json:"a_id"`
	A    *A       `db:"a" src:"a_id" dest:"id" table:"public.a" json:"a,omitempty"`
}
type D struct {
	_    struct{}  `db:"d" json:"-"`
	ID   string    `db:"id" json:"id"`
	Date time.Time `db:"date" json:"date"`
	As   []*A      `db:"as" src:"id" dest:"id" table:"public.a" through:"public.ad" through_src:"d_id" through_dest:"a_id" json:"as,omitempty"`
}
type AD struct {
	_   struct{} `db:"ad" json:"-"`
	AID string   `db:"a_id" json:"a_id"`
	DID string   `db:"d_id" json:"d_id"`
}

var (
	ABuilder = repository.NewSQLBuilder[A](
		repository.UuidV7Generator,
	)
	BBuilder = repository.NewSQLBuilder[B](
		repository.UuidV7Generator,
	)
	CBuilder = repository.NewSQLBuilder[C](
		repository.UuidV7Generator,
	)
	DBuilder = repository.NewSQLBuilder[D](
		repository.UuidV7Generator,
	)
	ADBuilder = repository.NewSQLBuilder[AD](
		repository.InsertID,
	)
)

type builderWhereTest[T any] struct {
	name string // description of this test case
	// Named input parameters for target function.
	builder *repository.SQLBuilder[T]
	where   *map[string]any
	args    *[]any
	want    string
	wantErr bool
}

func TestSQLBuilder_Models_WhereError(t *testing.T) {

	tests := []builderWhereTest[A]{
		{
			name:    "a: where id_eq_hello",
			builder: ABuilder,
			where: &map[string]any{
				"id": map[string]any{
					"_eq": "hello",
				},
			},
			args:    &[]any{},
			wantErr: false,
			want:    "SELECT public.a.id,public.a.age FROM public.a WHERE public.a.id = $1",
		},
		{
			name:    "a: where age_eq_10",
			builder: ABuilder,
			where: &map[string]any{
				"age": map[string]any{
					"_eq": 10,
				},
			},
			args:    &[]any{},
			wantErr: false,
			want:    "SELECT public.a.id,public.a.age FROM public.a WHERE public.a.age = $1",
		},
		{
			name:    "a: where id_eq_hello, age_eq_10",
			builder: ABuilder,
			where: &map[string]any{
				"id": map[string]any{
					"_eq": "hello",
				},
				"age": map[string]any{
					"_eq": 10,
				},
			},
			args:    &[]any{},
			wantErr: false,
			want:    "SELECT public.a.id,public.a.age FROM public.a WHERE public.a.id = $1 AND public.a.age = $2",
		},
		{
			name:    "a: where id_eq_hello or age_eq_10",
			builder: ABuilder,
			where: &map[string]any{
				"_or": []map[string]any{
					{
						"id": map[string]any{
							"_eq": "hello",
						},
					},
					{
						"age": map[string]any{
							"_eq": 10,
						},
					},
				},
			},
			args:    &[]any{},
			wantErr: false,
			want:    "SELECT public.a.id,public.a.age FROM public.a WHERE (public.a.id = $1 OR public.a.age = $2)",
		},
		{
			name:    "a: where id _eq hello and age _eq 10",
			builder: ABuilder,
			where: &map[string]any{
				"_and": []map[string]any{
					{
						"id": map[string]any{
							"_eq": "hello",
						},
					},
					{
						"age": map[string]any{
							"_eq": 10,
						},
					},
				},
			},
			args:    &[]any{},
			wantErr: false,
			want:    "SELECT public.a.id,public.a.age FROM public.a WHERE (public.a.id = $1 AND public.a.age = $2)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWhere(t, tt)
		})
	}
}

func testWhere[A any](t *testing.T, tt builderWhereTest[A]) {
	query := fmt.Sprintf("SELECT %s FROM %s", tt.builder.QualifiedColumnNamesJoined(), tt.builder.TableName())
	got, gotErr := tt.builder.WhereError(context.Background(), tt.where, tt.args)
	query += fmt.Sprintf(" WHERE %s", got)
	if gotErr != nil {
		if !tt.wantErr {
			t.Errorf("WhereError() failed: %v", gotErr)
		}
	}
	if tt.wantErr {
		t.Fatal("WhereError() succeeded unexpectedly")
	}
	if query != tt.want {
		t.Errorf("WhereError() = %v, want %v", query, tt.want)
	}
}
