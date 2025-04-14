package shared

const (
	StripeProductProID      string = "prod_pro"
	StripeProductAdvancedID string = "prod_advanced"
)

var StripeRoleMap = map[string]string{
	StripeProductProID:      PermissionNamePro,
	StripeProductAdvancedID: PermissionNameAdvanced,
}
