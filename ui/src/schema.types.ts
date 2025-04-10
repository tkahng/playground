import { components } from "./schema";

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

export type UserDetail = {
  roles:
    | {
        readonly $schema?: string;
        created_at: string;
        description?: string;
        id: string;
        name: string;
        permissions?: components["schemas"]["Permission"][] | null;
        updated_at: string;
      }[]
    | null;
  permissions: {
    created_at: string;
    description?: string;
    id: string;
    is_directly_assigned: boolean;
    name: string;
    roles: {
      created_at: string;
      description?: string;
      id: string;
      name: string;
      permissions?: components["schemas"]["Permission"][] | null;
      updated_at: string;
    }[];
    updated_at: string;
  }[];
  $schema?: string;
  created_at: string;
  email: string;
  email_verified_at: string;
  id: string;
  image: string;
  name: string;
  updated_at: string;
};
