package repository_test

import "github.com/tkahng/playground/internal/database/repository"

type A struct {
	_  struct{} `db:"a" json:"-"`
	ID string   `db:"id" json:"id"`
	B  *B       `db:"b" src:"id" dest:"a_id" table:"b" json:"b,omitempty"`
	Cs []*C     `db:"cs" src:"id" dest:"a_id" table:"c" json:"cs,omitempty"`
	Ds []*D     `db:"ds" src:"id" dest:"id" table:"d" through:"ad" through_src:"a_id" through_dest:"d_id" json:"ds,omitempty"`
}
type B struct {
	_   struct{} `db:"public.b" json:"-"`
	ID  string   `db:"id" json:"id"`
	AID string   `db:"a_id" json:"a_id"`
	A   *A       `db:"a" src:"a_id" dest:"id" table:"a" json:"a,omitempty"`
}
type C struct {
	_   struct{} `db:"public.b" json:"-"`
	ID  string   `db:"id" json:"id"`
	AID string   `db:"a_id" json:"a_id"`
	A   *A       `db:"a" src:"a_id" dest:"id" table:"a" json:"a,omitempty"`
}
type D struct {
	_  struct{} `db:"public.b" json:"-"`
	ID string   `db:"id" json:"id"`
	As []*A     `db:"as" src:"id" dest:"id" table:"a" through:"ad" through_src:"d_id" through_dest:"a_id" json:"as,omitempty"`
}
type AD struct {
	_   struct{} `db:"public.b" json:"-"`
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
