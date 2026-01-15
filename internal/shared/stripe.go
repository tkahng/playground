package shared

const (
	StripeProductProID      string = "prod_pro"
	StripeProductAdvancedID string = "prod_advanced"
)

var StripeProductPermissionMap = map[string]string{
	StripeProductProID:      PermissionNamePro,
	StripeProductAdvancedID: PermissionNameAdvanced,
}
