import {
  AuthenticatedDTO,
  RefreshTokenInput,
  SigninInput,
  SignupInput,
  User,
} from "@/schema.types";
import { client } from "./client";

export const signIn = async (
  args: SigninInput
): Promise<AuthenticatedDTO | null> => {
  const {
    data,
    error,
    // response: { status },
  } = await client.POST("/api/auth/signin", {
    body: {
      email: args.email,
      password: args.password,
    },
  });
  if (error) {
    throw new Error(error.detail);
  }
  return data || null;
};

export const refreshToken = async (
  args: RefreshTokenInput
): Promise<AuthenticatedDTO | null> => {
  const {
    data,
    error,
    // response: { status },
  } = await client.POST("/api/auth/refresh-token", {
    body: {
      refresh_token: args.refresh_token,
    },
  });
  if (error) {
    throw new Error(error.detail);
  }
  return data;
};

export const signUp = async (
  args: SignupInput
): Promise<AuthenticatedDTO | null> => {
  const {
    data,
    error,
    // response: { status },
  } = await client.POST("/api/auth/signup", {
    body: args,
  });
  if (error) {
    throw new Error(error.detail);
  }
  return data;
};

export const getMe = async (token: string): Promise<User> => {
  const { data, error } = await client.GET("/api/auth/me", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw new Error(error.detail);
  }
  return data;
};

type UserPaginate = {
  page?: number;
  per_page?: number;
  providers?:
    | ("google" | "apple" | "facebook" | "github" | "credentials")[]
    | null;
  q?: string;
  ids?: string[] | null;
  emails?: string[] | null;
  role_ids?: string[] | null;
  permission_ids?: string[] | null;
  sort_by?: string;
  sort_order?: string;
  expand?: ("roles" | "permissions" | "accounts" | "subscriptions")[] | null;
};

export const userPaginate = async (token: string, args: UserPaginate) => {
  const { data, error } = await client.GET("/api/admin/users", {
    params: {
      query: args,
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (error) {
    throw new Error(error.detail);
  }
  return data;
};
