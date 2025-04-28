package shared

const (
	StripeProductProID      string = "prod_pro"
	StripeProductAdvancedID string = "prod_advanced"
)

var StripeRoleMap = map[string]string{
	StripeProductProID:      PermissionNamePro,
	StripeProductAdvancedID: PermissionNameAdvanced,
}

type ActiveStatus string

const (
	Active   ActiveStatus = "active"
	Inactive ActiveStatus = "inactive"
)

type StripeCustomerListFilter struct {
	Q   string   `query:"q,omitempty" required:"false"`
	Ids []string `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
}
type StripeCustomerListParams struct {
	PaginatedInput
	StripeCustomerListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"users"`
}
