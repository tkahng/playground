import { components, operations } from "@/schema";

export type SigninInput = components["schemas"]["SigninDto"];

export type UserInfoTokens = components["schemas"]["ApiUserInfoTokens"];
export type SignupInput = components["schemas"]["SignupInput"];

export type RefreshTokenInput = components["schemas"]["RefreshTokenInput"];

export type User = components["schemas"]["ApiUser"];

export type PriceIntervals = components["schemas"]["StripePrice"]["interval"];

export type BillingIntervals = Exclude<
  PriceIntervals,
  "week" | "day" | undefined
>;

export type RoleWithPermissions = components["schemas"]["Role"];
export type Role = components["schemas"]["Role"];
export type Permission = components["schemas"]["Permission"];

export type UserDetail = components["schemas"]["ApiUser"];

export type UserDetailWithRoles = {
  accounts: components["schemas"]["UserAccountOutput"][];
  roles: RoleWithPermissions[] | null;
  permissions: {
    created_at: string;
    description?: string | null;
    id: string;
    is_directly_assigned: boolean;
    name: string;
    roles: Role[];
    updated_at: string;
  }[];
  $schema?: string;
  created_at: string;
  email: string;
  email_verified_at: string | null;
  id: string;
  image: string | null;
  name: string | null;
  updated_at: string;
};

export type StripeSubscription = components["schemas"]["StripeSubscription"];

export type SubscriptionWithPrice = StripeSubscription;

export type ProductWithPrices = components["schemas"]["StripeProduct"];

export type Price = components["schemas"]["StripePrice"];

export type UserPermissions = components["schemas"]["PermissionSource"];

export type ErrorModel = components["schemas"]["ErrorModel"];

export type TaskStatus = components["schemas"]["Task"]["status"];

export type UserWithAccounts = components["schemas"]["UserWithAccounts"];

export type TeamMember = components["schemas"]["TeamMember"];

export type TeamMemberRole = components["schemas"]["TeamMember"]["role"];
export type Team = components["schemas"]["Team"];

export type TeamWithMember = components["schemas"]["TeamWithMember"];
export type TeamMemberState = {
  currentMember: TeamMember | null;
  members: TeamMember[];
};

export type TeamMemberNotificationData =
  components["schemas"]["NotificationPayloadNewTeamMemberNotificationData"];

export type JobsParams = operations["admin-jobs-get"]["parameters"]["query"];

export type TaskCreateParams =
  operations["task-project-tasks-create"]["requestBody"]["content"]["application/json"];

export type Task = components["schemas"]["Task"];
