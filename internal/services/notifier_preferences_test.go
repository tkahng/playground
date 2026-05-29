package services_test

// Tests for notification preference filtering in sendToMembers:
//
//   - Members with a disabled preference for the notification type are skipped.
//   - Members with no preference row (default) still receive notifications.
//   - Members with an enabled preference row receive notifications.
//   - UpsertNotificationPreference is idempotent (re-enable after disable).

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
)

func setupPreferenceFixture(t *testing.T, ctx context.Context, db database.Dbx) (
	adapter stores.StorageAdapterInterface,
	optedOut *models.TeamMember,
	optedIn *models.TeamMember,
	noPreference *models.TeamMember,
	teamID interface{ String() string },
) {
	t.Helper()
	adapter = stores.NewStorageAdapter(db)

	ownerUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "pref-owner@example.com"})
	require.NoError(t, err)
	owner, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, ownerUser, "pref-team", models.TeamMemberRoleOwner)
	require.NoError(t, err)

	optedOutUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "pref-optout@example.com"})
	require.NoError(t, err)
	optedOut, err = adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
		TeamID: owner.TeamID, UserID: &optedOutUser.ID, Role: models.TeamMemberRoleMember, Active: true,
	})
	require.NoError(t, err)

	optedInUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "pref-optin@example.com"})
	require.NoError(t, err)
	optedIn, err = adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
		TeamID: owner.TeamID, UserID: &optedInUser.ID, Role: models.TeamMemberRoleMember, Active: true,
	})
	require.NoError(t, err)

	noPrefUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "pref-none@example.com"})
	require.NoError(t, err)
	noPreference, err = adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
		TeamID: owner.TeamID, UserID: &noPrefUser.ID, Role: models.TeamMemberRoleMember, Active: true,
	})
	require.NoError(t, err)

	return adapter, optedOut, optedIn, noPreference, owner.TeamID
}

// TestPreferences_DisabledMemberSkipped verifies that a member who has opted out
// of a notification type does not receive a notification.
func TestPreferences_DisabledMemberSkipped(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, optedOut, _, noPreference, _ := setupPreferenceFixture(t, ctx, db)

		require.NoError(t, adapter.Notification().UpsertNotificationPreference(ctx, optedOut.ID, "assigned_to_task", false))

		project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name: "pref-project", Status: models.TaskProjectStatusTodo,
			TeamID: optedOut.TeamID, MemberID: noPreference.ID,
		})
		require.NoError(t, err)

		assignerUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "pref-assigner@example.com"})
		require.NoError(t, err)
		assigner, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID: optedOut.TeamID, UserID: &assignerUser.ID, Role: models.TeamMemberRoleMember, Active: true,
		})
		require.NoError(t, err)

		task, err := adapter.Task().CreateTask(ctx, &models.Task{
			Name: "pref-task", Status: models.TaskStatusTodo,
			TeamID: optedOut.TeamID, ProjectID: project.ID, AssigneeID: &optedOut.ID,
		})
		require.NoError(t, err)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		require.NoError(t, notifier.NotifyAssignedToTask(ctx, task.ID, assigner.ID, optedOut.ID))

		count, err := adapter.Notification().CountNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{optedOut.ID},
			Types:         []string{"assigned_to_task"},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "opted-out member should not receive notification")
	})
}

// TestPreferences_DefaultEnabled verifies that a member with no preference row
// (the default state) still receives notifications.
func TestPreferences_DefaultEnabled(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, _, _, noPreference, _ := setupPreferenceFixture(t, ctx, db)

		assignerUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "pref-assigner2@example.com"})
		require.NoError(t, err)
		assigner, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID: noPreference.TeamID, UserID: &assignerUser.ID, Role: models.TeamMemberRoleMember, Active: true,
		})
		require.NoError(t, err)

		project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name: "pref-project2", Status: models.TaskProjectStatusTodo,
			TeamID: noPreference.TeamID, MemberID: assigner.ID,
		})
		require.NoError(t, err)
		task, err := adapter.Task().CreateTask(ctx, &models.Task{
			Name: "pref-task2", Status: models.TaskStatusTodo,
			TeamID: noPreference.TeamID, ProjectID: project.ID, AssigneeID: &noPreference.ID,
		})
		require.NoError(t, err)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		require.NoError(t, notifier.NotifyAssignedToTask(ctx, task.ID, assigner.ID, noPreference.ID))

		count, err := adapter.Notification().CountNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{noPreference.ID},
			Types:         []string{"assigned_to_task"},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "member with no preference should receive notification by default")
	})
}

// TestPreferences_ExplicitlyEnabled verifies that a member who has explicitly
// opted in still receives notifications.
func TestPreferences_ExplicitlyEnabled(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, _, optedIn, _, _ := setupPreferenceFixture(t, ctx, db)

		require.NoError(t, adapter.Notification().UpsertNotificationPreference(ctx, optedIn.ID, "assigned_to_task", true))

		assignerUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "pref-assigner3@example.com"})
		require.NoError(t, err)
		assigner, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID: optedIn.TeamID, UserID: &assignerUser.ID, Role: models.TeamMemberRoleMember, Active: true,
		})
		require.NoError(t, err)

		project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name: "pref-project3", Status: models.TaskProjectStatusTodo,
			TeamID: optedIn.TeamID, MemberID: assigner.ID,
		})
		require.NoError(t, err)
		task, err := adapter.Task().CreateTask(ctx, &models.Task{
			Name: "pref-task3", Status: models.TaskStatusTodo,
			TeamID: optedIn.TeamID, ProjectID: project.ID, AssigneeID: &optedIn.ID,
		})
		require.NoError(t, err)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		require.NoError(t, notifier.NotifyAssignedToTask(ctx, task.ID, assigner.ID, optedIn.ID))

		count, err := adapter.Notification().CountNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{optedIn.ID},
			Types:         []string{"assigned_to_task"},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "member with enabled preference should receive notification")
	})
}

// TestPreferences_FailClosedOnDbError verifies that a database error while loading
// preferences causes sendToMembers to return an error instead of notifying everyone.
func TestPreferences_FailClosedOnDbError(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		// Inject a failing FindDisabledMemberIDs — simulates a DB outage.
		adapter.NotificationFunc.FindDisabledMemberIDsFunc = func(_ context.Context, _ []uuid.UUID, _ string) ([]uuid.UUID, error) {
			return nil, errors.New("db connection lost")
		}

		_, optedOut, _, noPreference, _ := setupPreferenceFixture(t, ctx, db)

		assignerUser, err := adapter.User().CreateUser(ctx, &models.User{Email: "failclosed-assigner@example.com"})
		require.NoError(t, err)
		assigner, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID: optedOut.TeamID, UserID: &assignerUser.ID, Role: models.TeamMemberRoleMember, Active: true,
		})
		require.NoError(t, err)

		project, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name: "failclosed-project", Status: models.TaskProjectStatusTodo,
			TeamID: optedOut.TeamID, MemberID: assigner.ID,
		})
		require.NoError(t, err)
		task, err := adapter.Task().CreateTask(ctx, &models.Task{
			Name: "failclosed-task", Status: models.TaskStatusTodo,
			TeamID: optedOut.TeamID, ProjectID: project.ID, AssigneeID: &noPreference.ID,
		})
		require.NoError(t, err)

		notifier := services.NewDbNotificationPublisher(noopSSE{}, services.NewTeamService(adapter), adapter)
		err = notifier.NotifyAssignedToTask(ctx, task.ID, assigner.ID, noPreference.ID)
		require.Error(t, err, "preference query failure must propagate — not silently notify everyone")

		count, err := adapter.Notification().CountNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{noPreference.ID},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "no notifications should be persisted when preference query fails")
	})
}

// TestPreferences_UpsertIsIdempotent verifies that upserting the same preference
// multiple times does not create duplicate rows or cause errors.
func TestPreferences_UpsertIsIdempotent(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, optedOut, _, _, _ := setupPreferenceFixture(t, ctx, db)

		require.NoError(t, adapter.Notification().UpsertNotificationPreference(ctx, optedOut.ID, "task_completed", false))
		require.NoError(t, adapter.Notification().UpsertNotificationPreference(ctx, optedOut.ID, "task_completed", false))
		// re-enable
		require.NoError(t, adapter.Notification().UpsertNotificationPreference(ctx, optedOut.ID, "task_completed", true))

		disabled, err := adapter.Notification().FindDisabledMemberIDs(ctx, []uuid.UUID{optedOut.ID}, "task_completed")
		require.NoError(t, err)
		assert.Empty(t, disabled, "member should be re-enabled after upsert with enabled=true")
	})
}
