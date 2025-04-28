package shared

const (
	StripeProductProID      string = "prod_pro"
	StripeProductAdvancedID string = "prod_advanced"
)

var StripeRoleMap = map[string]string{
	StripeProductProID:      PermissionNamePro,
	StripeProductAdvancedID: PermissionNameAdvanced,
}

type StripeProductListFilter struct {
	Q      string       `query:"q,omitempty" required:"false"`
	Ids    []string     `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Active ActiveStatus `query:"active,omitempty" required:"false" enum:"active,inactive"`
}

type StripeProductListParams struct {
	PaginatedInput
	StripeProductListFilter
	SortParams
	Expand      []string     `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"prices,roles"`
	PriceActive ActiveStatus `query:"price_active,omitempty" required:"false" enum:"active,inactive"`
}

type ActiveStatus string

const (
	Active   ActiveStatus = "active"
	Inactive ActiveStatus = "inactive"
)

type StripePriceListFilter struct {
	Q      string       `query:"q,omitempty" required:"false"`
	Ids    []string     `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Active ActiveStatus `query:"active,omitempty" required:"false" enum:"active,inactive"`
}
type StripePriceListParams struct {
	PaginatedInput
	StripePriceListFilter
	SortParams
}

type StripeCustomerListFilter struct {
	Q   string   `query:"q,omitempty" required:"false"`
	Ids []string `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
}
type StripeCustomerListParams struct {
	PaginatedInput
	StripeCustomerListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"users"`
}

type StripeSubscriptionListFilter struct {
	Q      string                     `query:"q,omitempty" required:"false"`
	Ids    []string                   `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	UserID string                     `query:"user_id,omitempty" required:"false" format:"uuid"`
	Status []StripeSubscriptionStatus `query:"status,omitempty" required:"false" minimum:"1" maximum:"100" enum:"trialing,active,canceled,incomplete,incomplete_expired,past_due,unpaid,paused"`
}
type StripeSubscriptionListParams struct {
	PaginatedInput
	StripeSubscriptionListFilter
	SortParams
	StripeSubscriptionExpand
	// Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"user,price,product"`
}

type StripeSubscriptionExpand struct {
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"user,price,product"`
}
