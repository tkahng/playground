package shared

type UserPermissionsListFilter struct {
	UserId  string `path:"user-id" format:"uuid"`
	Reverse bool   `query:"reverse,omitempty"`
}
type UserPermissionsListParams struct {
	PaginatedInput
	UserPermissionsListFilter
	SortParams
}
