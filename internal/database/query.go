package database

type QueryOptions struct {
	tx Dbx
}

type Queryer interface {
	Db() Dbx
	SetTx(Dbx)
}
