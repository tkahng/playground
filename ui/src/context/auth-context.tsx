import { useLocalStorage } from "@/hooks/use-local-storage";
import { client } from "@/lib/client";
import { AuthenticatedDTO, SigninInput, SignupInput } from "@/schema.types";
import { jwtDecode } from "jwt-decode";
import { createContext, useMemo } from "react";

interface AuthProviderProps {
  children?: React.ReactNode;
}

export interface AuthContextType {
  // client():
  user: AuthenticatedDTO | null;
  signUp: ({ email, name, password }: SignupInput) => Promise<void>;
  login: ({ email, password }: SigninInput) => Promise<any>;
  logout: () => Promise<void>;
  checkError: (error: any) => Promise<void>;
  checkAuth: () => Promise<void>;
  getOrRefreshToken: (token?: string) => Promise<AuthenticatedDTO>;
}
export const AuthContext = createContext<AuthContextType>({
  user: null,
  signUp: async () => {},
  login: async () => {},
  logout: async () => {},
  checkError: async () => {},
  checkAuth: async () => {},
  getOrRefreshToken: async () => {
    throw new Error("Not implemented");
  },
});

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [user, setUser] = useLocalStorage<AuthenticatedDTO | null>(
    "auth",
    null
  );
  const signUp = async ({
    email,
    name,
    password,
  }: SignupInput): Promise<void> => {
    const { data, error } = await client.POST("/api/auth/signup", {
      body: {
        email,
        name,
        password,
      },
    });
    if (error) {
      throw error;
    }
    setUser(data);
    return Promise.resolve();
  };
  const login = async ({ email, password }: SigninInput): Promise<void> => {
    const { data, error } = await client.POST("/api/auth/signin", {
      body: {
        email,
        password,
      },
    });
    if (error) {
      throw error;
    }
    setUser(data);
    return Promise.resolve();
  };

  const logout = async (): Promise<void> => {
    setUser(null);
    return Promise.resolve();
  };

  const checkError = async (error: any) => {
    if (
      error.status === 401 ||
      error.status === 403 ||
      // Supabase returns 400 when the session is missing, we need to check this case too.
      (error.status === 400 && error.name === "AuthSessionMissingError")
    ) {
      return Promise.reject();
    }

    return Promise.resolve();
  };

  const getOrRefreshToken = async (token?: string) => {
    if (token) {
      const { data, error } = await client.POST("/api/auth/refresh-token", {
        body: {
          refresh_token: token,
        },
      });
      if (error) {
        throw error;
      }
      setUser(data);
      return data;
    }
    if (!user) {
      return Promise.reject();
    } else {
      const decoded = jwtDecode(user.tokens.access_token) as any;
      if (decoded.exp <= Math.round(Date.now() / 1000)) {
        const { data, error } = await client.POST("/api/auth/refresh-token", {
          body: {
            refresh_token: user.tokens.refresh_token,
          },
        });
        if (error) {
          setUser(null);
          throw error;
        }
        setUser(data);
        return user;
      } else {
        return user;
      }
    }
  };

  const checkAuth = async () => {
    await getOrRefreshToken();
  };

  const value = useMemo(
    () => ({
      user,
      signUp,
      login,
      logout,
      checkAuth,
      checkError,
      getOrRefreshToken,
    }),
    [user]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
