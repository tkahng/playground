import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, type RenderOptions } from "@testing-library/react";
import React from "react";
import { AuthContext, type AuthContextType } from "@/context/auth-context";
import type { UserInfoTokens } from "@/schema.types";

export const mockUserTokens: UserInfoTokens = {
  user: {
    id: "user-1",
    email: "test@example.com",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  } as UserInfoTokens["user"],
  tokens: {
    access_token: "test-access-token",
    refresh_token: "test-refresh-token",
  },
};

export const mockAuthContext: AuthContextType = {
  user: mockUserTokens,
  signUp: vi.fn(),
  login: vi.fn(),
  logout: vi.fn(),
  checkAuth: vi.fn(),
  getOrRefreshToken: vi.fn(),
};

export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  });
}

function AllProviders({ children }: { children: React.ReactNode }) {
  const queryClient = createTestQueryClient();
  return (
    <QueryClientProvider client={queryClient}>
      <AuthContext.Provider value={mockAuthContext}>
        {children}
      </AuthContext.Provider>
    </QueryClientProvider>
  );
}

function customRender(ui: React.ReactElement, options?: RenderOptions) {
  return render(ui, { wrapper: AllProviders, ...options });
}

export { customRender as render };
export * from "@testing-library/react";
