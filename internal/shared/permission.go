package shared

import "github.com/tkahng/playground/internal/models"

const (
	PermissionNameAdmin    string = "superuser"
	PermissionNameBasic    string = "basic"
	PermissionNamePro      string = "pro"
	PermissionNameAdvanced string = "advanced"

	TeamPermissionSettingsManage string = "team.settings.manage"
	TeamPermissionDelete         string = "team.delete"
	TeamPermissionMembersInvite  string = "team.members.invite"
	TeamPermissionMembersManage  string = "team.members.manage"
	TeamPermissionBillingManage  string = "team.billing.manage"
	TeamPermissionProjectsCreate string = "projects.create"
	TeamPermissionProjectsManage string = "projects.manage"
	TeamPermissionProjectsDelete string = "projects.delete"
	TeamPermissionTasksCreate    string = "tasks.create"
	TeamPermissionTasksEdit      string = "tasks.edit"
	TeamPermissionTasksAssign    string = "tasks.assign"
	TeamPermissionTasksDelete    string = "tasks.delete"
	TeamPermissionWorkflowManage string = "workflow.manage"

)

var (
	KnownRoleNames, KnwonPermissionNames                     = []string{PermissionNameAdmin, PermissionNameAdvanced, PermissionNamePro, PermissionNameBasic}, []string{PermissionNameAdmin, PermissionNameAdvanced, PermissionNamePro, PermissionNameBasic}
	KnownRoleNamesPermissionsMap         map[string][]string = map[string][]string{
		PermissionNameBasic:    {PermissionNameBasic},
		PermissionNamePro:      {PermissionNameBasic, PermissionNamePro},
		PermissionNameAdvanced: {PermissionNameBasic, PermissionNamePro, PermissionNameAdvanced},
		PermissionNameAdmin:    {PermissionNameBasic, PermissionNamePro, PermissionNameAdvanced, PermissionNameAdmin},
	}
	KnownTeamRolePermissionsMap map[string][]string = map[string][]string{
		string(models.TeamMemberRoleOwner): {
			TeamPermissionSettingsManage,
			TeamPermissionDelete,
			TeamPermissionMembersInvite,
			TeamPermissionMembersManage,
			TeamPermissionBillingManage,
			TeamPermissionProjectsCreate,
			TeamPermissionProjectsManage,
			TeamPermissionProjectsDelete,
			TeamPermissionTasksCreate,
			TeamPermissionTasksEdit,
			TeamPermissionTasksAssign,
			TeamPermissionTasksDelete,
			TeamPermissionWorkflowManage,
		},
		string(models.TeamMemberRoleMember): {
			TeamPermissionProjectsCreate,
			TeamPermissionTasksCreate,
			TeamPermissionTasksEdit,
			TeamPermissionTasksAssign,
		},
	}
)
