import { components } from "@/schema";

export type SigninInput = components["schemas"]["SigninDto"];

export type AuthenticatedDTO = components["schemas"]["AuthenticatedDTO"];
export type SignupInput = components["schemas"]["SignupInput"];

export type RefreshTokenInput = components["schemas"]["RefreshTokenInput"];
export type RefreshTokenOutput = components["schemas"]["AuthenticatedDTO"];

export type User = components["schemas"]["User"];

export type PriceIntervals = components["schemas"]["Price"]["interval"];

export type BillingIntervals = Exclude<
  PriceIntervals,
  "week" | "day" | undefined
>;

export type UserInfo = components["schemas"]["UserDetail"];

export type RoleWithPermissions = components["schemas"]["RoleWithPermissions"];
export type Role = components["schemas"]["Role"];
export type Permission = components["schemas"]["Permission"];

export type UserDetail = components["schemas"]["UserDetail"];

export type UserDetailWithRoles = {
  roles: RoleWithPermissions[] | null;
  permissions: {
    created_at: string;
    description?: string;
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

export type SubscriptionWithPrice =
  components["schemas"]["SubscriptionWithPrice"];

export type ProductWithPrices =
  components["schemas"]["StripeProductWithPrices"];

export type Price = components["schemas"]["Price"];

export type UserPermissions = components["schemas"]["PermissionSource"];

export type ErrorModel = components["schemas"]["ErrorModel"];
