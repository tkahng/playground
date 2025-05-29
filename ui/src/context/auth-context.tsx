import { refreshToken, signIn } from "@/lib/queries";
import { SigninInput, SignupInput, UserInfoTokens } from "@/schema.types";
import { jwtDecode } from "jwt-decode";
import React from "react";

export interface AuthContextType {
  user: UserInfoTokens | null;
  setUser: (user: UserInfoTokens | null) => void;
  signUp: (args: SignupInput) => Promise<UserInfoTokens>;
  login: (args: SigninInput) => Promise<UserInfoTokens>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
  getOrRefreshToken: (token?: string) => Promise<UserInfoTokens>;
}

export const AuthContext = React.createContext<AuthContextType>({
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

export const AuthProvider: React.FC<{
  children: React.ReactNode;
}> = ({ children }) => {
  const [user, setUser] = React.useState<UserInfoTokens | null>(null);
  const values = React.useMemo(() => {
    const signUp = async (args: SignupInput): Promise<UserInfoTokens> => {
      const data = await signIn(args);
      setUser(data);
      return data;
    };
    const login = async (args: SigninInput): Promise<UserInfoTokens> => {
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
          const decoded = jwtDecode(user.tokens.access_token);
          if (!decoded?.exp) {
            console.error("Token does not have an expiration time.");
            return Promise.reject();
          }
          if (decoded?.exp <= Math.round(Date.now() / 1000)) {
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

  return <AuthContext.Provider value={values}>{children}</AuthContext.Provider>;
};
