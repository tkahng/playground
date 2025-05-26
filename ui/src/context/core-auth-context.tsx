// react context for authentication

import { refreshToken, signIn } from "@/lib/queries";
import { components } from "@/schema";
import { SigninInput, SignupInput, UserInfoTokens } from "@/schema.types";
import { jwtDecode } from "jwt-decode";
import React from "react";

interface DecodedToken {
  exp: number; // Expiration time in seconds since the epoch
}

export interface CoreAuthContextType {
  user: UserInfoTokens | null;
  setUser: (user: UserInfoTokens | null) => void;
  signUp: (args: SignupInput) => Promise<UserInfoTokens>;
  login: (args: SigninInput) => Promise<UserInfoTokens>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
  getOrRefreshToken: (token?: string) => Promise<UserInfoTokens>;
}

export const CoreAuthContext = React.createContext<CoreAuthContextType>({
  user: null,
  setUser: () => {},
  signUp: async () => {
    throw new Error("Not implemented");
  },
  login: async () => {
    throw new Error("Not implemented");
  },
  logout: async () => {
    throw new Error("Not implemented");
  },
  getOrRefreshToken: async () => {
    throw new Error("Not implemented");
  },
  checkAuth: async () => {
    throw new Error("Not implemented");
  },
});

export const CoreAuthProvider: React.FC<{
  children: React.ReactNode;
}> = ({ children }) => {
  const [user, setUser] = React.useState<UserInfoTokens | null>(null);
  const signUp = async (args: SignupInput): Promise<UserInfoTokens> => {
    const data = await signIn(args);
    setUser(data);
    return data;
  };
  const values = React.useMemo(() => {
    const login = async (
      args: components["schemas"]["SigninDto"]
    ): Promise<UserInfoTokens> => {
      const data = await signIn(args);
      setUser(data);
      return data;
    };
    const logout = async () => {
      setUser(null);
    };
    const getOrRefreshToken = async (token?: string) => {
      try {
        if (token) {
          const data = await refreshToken({ refresh_token: token });
          setUser(data);
          return data;
        }
        if (!user) {
          return Promise.reject();
        } else {
          const decoded = jwtDecode<DecodedToken>(user.tokens.access_token);
          if (decoded.exp <= Math.round(Date.now() / 1000)) {
            const data = await refreshToken({
              refresh_token: user.tokens.refresh_token,
            });
            setUser(data);
            return data;
          } else {
            return user;
          }
        }
      } catch (error) {
        console.error("Error refreshing token:", error);
        setUser(null);
        return Promise.reject();
      }
    };
    const checkAuth = async () => {
      if (!user) {
        return;
      }
      try {
        await getOrRefreshToken(user.tokens.refresh_token);
      } catch (error) {
        console.error("Error checking auth:", error);
        setUser(null);
        return Promise.reject();
      }
    };
    return {
      user,
      signUp,
      setUser,
      login,
      logout,
      getOrRefreshToken,
      checkAuth,
    };
  }, [user]);

  return (
    <CoreAuthContext.Provider value={values}>
      {children}
    </CoreAuthContext.Provider>
  );
};
