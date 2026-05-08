import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  RouterContextProvider,
  createMemoryHistory,
  createRootRoute,
  createRouter,
  Outlet,
} from "@tanstack/react-router";
import { render, type RenderOptions } from "@testing-library/react";
import React from "react";
import { AuthContext, type AuthContextType } from "@/context/auth-context";
import type { UserInfoTokens } from "@/schema.types";
import { vi } from "vitest";

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
    expires_in: 0,
    token_type: "",
  },
  permissions: null,
  providers: null,
  roles: null,
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

const testRootRoute = createRootRoute({ component: Outlet });
const testRouteTree = testRootRoute.addChildren([]);
const testRouter = createRouter({
  routeTree: testRouteTree,
  history: createMemoryHistory({ initialEntries: ["/"] }),
});

function AllProviders({ children }: { children: React.ReactNode }) {
  const queryClient = createTestQueryClient();
  return (
    <QueryClientProvider client={queryClient}>
      <AuthContext.Provider value={mockAuthContext}>
        <RouterContextProvider router={testRouter}>
          {children}
        </RouterContextProvider>
      </AuthContext.Provider>
    </QueryClientProvider>
  );
}

function customRender(ui: React.ReactElement, options?: RenderOptions) {
  return render(ui, { wrapper: AllProviders, ...options });
}

export { customRender as render };
export * from "@testing-library/react";
