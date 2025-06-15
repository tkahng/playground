package shared

const (
	StripeProductBasicID    string = "prod_basic"
	StripeProductProID      string = "prod_pro"
	StripeProductAdvancedID string = "prod_advanced"
)

var StripeProductPermissionMap = map[string]string{
	StripeProductBasicID:    PermissionNameBasic,
	StripeProductProID:      PermissionNamePro,
	StripeProductAdvancedID: PermissionNameAdvanced,
}
