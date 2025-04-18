package shared

type UserPermissionsListFilter struct {
	UserId  string `path:"userId" format:"uuid"`
	Reverse bool   `query:"reverse,omitempty"`
}
type UserPermissionsListParams struct {
	PaginatedInput
	UserPermissionsListFilter
	SortParams
}
