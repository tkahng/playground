import React from "react";
import { render } from "@testing-library/react";
import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  vi,
} from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/context/auth-context";
import { mockAuthContext, mockUserTokens } from "@/test/test-utils";

// ─── Module mocks (hoisted by vitest) ───────────────────────────────────────

vi.mock("@/lib/api", () => ({
  createLedgerWallet: vi.fn().mockResolvedValue(undefined),
}));

const navigateSpy = vi.fn();

// Custom Navigate that records props when rendered — lets us assert redirect
// destination and search params without running TanStack Router's full pipeline.
function FakeNavigate(props: { to: string; search?: Record<string, string> }) {
  navigateSpy(props);
  return null;
}

const outletSpy = vi.fn();
function FakeOutlet() {
  outletSpy();
  return null;
}

vi.mock("@tanstack/react-router", async () => {
  const actual =
    await vi.importActual<typeof import("@tanstack/react-router")>(
      "@tanstack/react-router"
    );
  return {
    ...actual,
    Navigate: FakeNavigate,
    Outlet: FakeOutlet,
    useRouterState: vi.fn(),
    createFileRoute: () => (opts: unknown) => opts, // no-op in tests
  };
});

// ─── Import subject AFTER mocks are in place ────────────────────────────────

// eslint-disable-next-line import/first
import { useRouterState } from "@tanstack/react-router";
// eslint-disable-next-line import/first
import { AuthGuard } from "@/routes/_authenticated";

const mockUseRouterState = vi.mocked(useRouterState);

// ─── Helpers ────────────────────────────────────────────────────────────────

function makeLocation(pathname: string, searchStr = "") {
  return {
    pathname,
    searchStr,
    search: Object.fromEntries(new URLSearchParams(searchStr)),
    hash: "",
    href: pathname + searchStr,
  };
}

function renderGuard(
  pathname: string,
  searchStr = "",
  authOverrides: Partial<typeof mockAuthContext> = {}
) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  mockUseRouterState.mockReturnValue(makeLocation(pathname, searchStr) as any);

  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const auth = { ...mockAuthContext, ...authOverrides };

  return render(
    <QueryClientProvider client={queryClient}>
      <AuthContext.Provider value={auth}>
        <AuthGuard />
      </AuthContext.Provider>
    </QueryClientProvider>
  );
}

// ─── Tests ──────────────────────────────────────────────────────────────────

describe("AuthGuard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("unauthenticated (user is null)", () => {
    it("renders Navigate to /signin for a generic protected path", () => {
      renderGuard("/dashboard", "", {
        user: null,
        getOrRefreshToken: vi.fn().mockRejectedValue(new Error()),
      });

      expect(navigateSpy).toHaveBeenCalledWith(
        expect.objectContaining({ to: "/signin" })
      );
    });

    it("puts the current pathname in redirect_to", () => {
      renderGuard("/dashboard", "", {
        user: null,
        getOrRefreshToken: vi.fn().mockRejectedValue(new Error()),
      });

      expect(navigateSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          to: "/signin",
          search: expect.objectContaining({ redirect_to: "/dashboard" }),
        })
      );
    });

    it("appends query string to redirect_to", () => {
      renderGuard("/dashboard", "?view=list&page=2", {
        user: null,
        getOrRefreshToken: vi.fn().mockRejectedValue(new Error()),
      });

      expect(navigateSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          search: expect.objectContaining({
            redirect_to: "/dashboard?view=list&page=2",
          }),
        })
      );
    });

    it("navigates to /signup for /team-invitation paths", () => {
      renderGuard("/team-invitation", "?token=abc", {
        user: null,
        getOrRefreshToken: vi.fn().mockRejectedValue(new Error()),
      });

      expect(navigateSpy).toHaveBeenCalledWith(
        expect.objectContaining({ to: "/signup" })
      );
      // Must not redirect to signin
      expect(navigateSpy).not.toHaveBeenCalledWith(
        expect.objectContaining({ to: "/signin" })
      );
    });

    it("includes token query string in redirect_to for team-invitation", () => {
      renderGuard("/team-invitation", "?token=abc", {
        user: null,
        getOrRefreshToken: vi.fn().mockRejectedValue(new Error()),
      });

      expect(navigateSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          search: expect.objectContaining({
            redirect_to: "/team-invitation?token=abc",
          }),
        })
      );
    });
  });

  describe("authenticated (user present)", () => {
    it("renders Outlet — no Navigate — when user is set", () => {
      renderGuard("/dashboard", "", {
        user: mockUserTokens,
        getOrRefreshToken: vi.fn().mockResolvedValue(mockUserTokens),
      });

      expect(outletSpy).toHaveBeenCalled();
      expect(navigateSpy).not.toHaveBeenCalled();
    });

    it("calls getOrRefreshToken exactly once on mount", () => {
      const getOrRefreshToken = vi.fn().mockResolvedValue(mockUserTokens);

      renderGuard("/dashboard", "", {
        user: mockUserTokens,
        getOrRefreshToken,
      });

      expect(getOrRefreshToken).toHaveBeenCalledOnce();
    });
  });
});
