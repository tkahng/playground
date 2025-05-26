// import { useLocalStorage } from "@/hooks/use-local-storage";
// import { client } from "@/lib/client";
// import { SigninInput, SignupInput, UserInfoTokens } from "@/schema.types";
// import { jwtDecode } from "jwt-decode"; // Ensure correct import
// import { createContext, useMemo } from "react";

// // Define the structure of the decoded JWT token
// interface DecodedToken {
//   exp: number; // Expiration time in seconds since the epoch
// }

// interface AuthProviderProps {
//   children?: React.ReactNode;
// }

// export interface AuthContextType {
//   user: UserInfoTokens | null;
//   signUp: ({ email, name, password }: SignupInput) => Promise<UserInfoTokens>;
//   login: ({ email, password }: SigninInput) => Promise<UserInfoTokens>;
//   logout: () => Promise<void>;

//   // checkError: (error: any) => Promise<UserInfoTokens>;
//   checkAuth: () => Promise<UserInfoTokens>;
//   getOrRefreshToken: (token?: string) => Promise<UserInfoTokens>;
// }

// export const AuthContext = createContext<AuthContextType>({
//   user: null,
//   signUp: async () => {
//     throw new Error("Not implemented");
//   },
//   login: async () => {
//     throw new Error("Not implemented");
//   },
//   logout: async () => {},
//   // checkError: async () => {
//   //   throw new Error("Not implemented");
//   // },
//   checkAuth: async () => {
//     throw new Error("Not implemented");
//   },
//   getOrRefreshToken: async () => {
//     throw new Error("Not implemented");
//   },
// });

// export const AuthProvider = ({ children }: AuthProviderProps) => {
//   const [user, setUser] = useLocalStorage<UserInfoTokens | null>("auth", null);

//   const signUp = async ({
//     email,
//     name,
//     password,
//   }: SignupInput): Promise<UserInfoTokens> => {
//     const { data, error } = await client.POST("/api/auth/signup", {
//       body: { email, name, password },
//     });
//     if (error) {
//       throw error;
//     }
//     setUser(data);
//     return data;
//   };

//   const login = async ({
//     email,
//     password,
//   }: SigninInput): Promise<UserInfoTokens> => {
//     const { data, error } = await client.POST("/api/auth/signin", {
//       body: { email, password },
//     });
//     if (error) {
//       throw error;
//     }
//     setUser(data);
//     return data;
//   };

//   const logout = async (): Promise<void> => {
//     setUser(null);
//     return Promise.resolve();
//   };

//   const getOrRefreshToken = async (token?: string) => {
//     if (token) {
//       const { data, error } = await client.POST("/api/auth/refresh-token", {
//         body: { refresh_token: token },
//       });
//       if (error) {
//         console.error("Error refreshing token:", error);
//         setUser(null);
//         throw error;
//       }
//       setUser(data);
//       return data;
//     }

//     if (!user) {
//       return Promise.reject();
//     } else {
//       const decoded = jwtDecode<DecodedToken>(user.tokens.access_token);
//       if (decoded.exp <= Math.round(Date.now() / 1000)) {
//         const { data, error } = await client.POST("/api/auth/refresh-token", {
//           body: { refresh_token: user.tokens.refresh_token },
//         });
//         if (error) {
//           setUser(null);
//           throw error;
//         }
//         setUser(data);
//         return data;
//       } else {
//         return user;
//       }
//     }
//   };

//   const checkAuth = async () => {
//     if (!user) {
//       return;
//     }
//     return getOrRefreshToken(user.tokens.refresh_token);
//   };

//   const value = useMemo(
//     () => ({
//       user,
//       signUp,
//       login,
//       logout,
//       checkAuth,
//       getOrRefreshToken,
//     }),
//     [user]
//   );
//   return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
// };
// react context for authentication

import { refreshToken, signIn } from "@/lib/queries";
import { components } from "@/schema";
import { SigninInput, SignupInput, UserInfoTokens } from "@/schema.types";
import { jwtDecode } from "jwt-decode";
import React from "react";

interface DecodedToken {
  exp: number; // Expiration time in seconds since the epoch
}

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

  return <AuthContext.Provider value={values}>{children}</AuthContext.Provider>;
};
