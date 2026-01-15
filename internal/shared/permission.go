package shared

const (
	PermissionNameAdmin    string = "superuser"
	PermissionNameBasic    string = "basic"
	PermissionNamePro      string = "pro"
	PermissionNameAdvanced string = "advanced"
)

var (
	KnownRoleNames, KnwonPermissionNames                     = []string{PermissionNameAdmin, PermissionNameAdvanced, PermissionNamePro, PermissionNameBasic}, []string{PermissionNameAdmin, PermissionNameAdvanced, PermissionNamePro, PermissionNameBasic}
	KnownRoleNamesPermissionsMap         map[string][]string = map[string][]string{
		PermissionNameBasic:    {PermissionNameBasic},
		PermissionNamePro:      {PermissionNameBasic, PermissionNamePro},
		PermissionNameAdvanced: {PermissionNameBasic, PermissionNamePro, PermissionNameAdvanced},
		PermissionNameAdmin:    {PermissionNameBasic, PermissionNamePro, PermissionNameAdvanced, PermissionNameAdmin},
	}
)
