import {
  AuthenticatedDTO,
  RefreshTokenInput,
  SigninInput,
  SignupInput,
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
  console.log({ data, error });
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
  console.log({ data, error });
  return data || null;
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
  console.log({ data, error });
  return data || null;
};

// export const getMe = async (token: string): Promise<void> => {
//   const { data, error } = await client.GET("/api/auth/me", {
//     headers: {
//       Authorization: `Bearer ${token}`,
//     },
//   });
// };
