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

export type UserInfo = components["schemas"]["UserInfo"];
