/**
 * Smoke tests: verify key page components render without throwing.
 *
 * These are not exhaustive — they confirm that the TanStack Router migration
 * didn't break imports or crash the render phase on the most-visited pages.
 */

import { screen, waitFor } from "@testing-library/react";
import React from "react";
import { describe, expect, it, vi } from "vitest";
import {
  render as customRender,
  mockAuthContext,
} from "@/test/test-utils";
import { AuthContext } from "@/context/auth-context";

// ─── Top-level stubs (hoisted by vitest) ────────────────────────────────────

vi.mock("@tanstack/react-router", async () => {
  const actual =
    await vi.importActual<typeof import("@tanstack/react-router")>(
      "@tanstack/react-router"
    );
  return {
    ...actual,
    Navigate: () => null,
    Outlet: () => null,
    useNavigate: () => vi.fn(),
    useRouterState: vi.fn().mockReturnValue({
      pathname: "/",
      searchStr: "",
      search: {},
      hash: "",
      href: "/",
    }),
    useParams: vi.fn().mockReturnValue({}),
    createFileRoute: () => (opts: unknown) => opts,
  };
});

vi.mock("@react-nano/use-event-source", () => ({
  useEventSource: vi.fn().mockReturnValue([null, "closed"]),
  useEventSourceListener: vi.fn(),
}));

vi.mock("@/components/connections", () => ({
  ProviderConnectionForm: () => null,
  providerNames: [],
}));

vi.mock("@/lib/queries", () => ({
  useActiveSubscription: vi.fn().mockReturnValue({ data: null, isLoading: false }),
  useMeQuery: vi.fn().mockReturnValue({ data: null, isLoading: false, isError: false, error: null }),
}));

vi.mock("@/lib/team-queries", () => ({
  getUserTeamMembers: vi.fn().mockResolvedValue({
    data: [],
    meta: { total: 0, page: 0, per_page: 10 },
  }),
  getTeamBySlug: vi.fn().mockResolvedValue({ team: null, member: null }),
  createTeam: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
  Toaster: () => null,
}));

// ─── Helper: render with no authenticated user ───────────────────────────────

function renderUnauthenticated(ui: React.ReactElement) {
  return customRender(
    <AuthContext.Provider value={{ ...mockAuthContext, user: null }}>
      {ui}
    </AuthContext.Provider>
  );
}

// ─── Landing ─────────────────────────────────────────────────────────────────

describe("Landing page", () => {
  it("renders without crashing", async () => {
    const { default: Landing } = await import("@/pages/landing/landing");
    customRender(<Landing />);
    await waitFor(() => expect(document.body).toBeInTheDocument());
  });
});

describe("Features page", () => {
  it("renders feature sections", async () => {
    const { default: Features } = await import("@/pages/landing/features");
    customRender(<Features />);
    await waitFor(() => expect(document.body).toBeInTheDocument());
  });
});

describe("Pricing page", () => {
  it("renders without crashing", async () => {
    const { default: PricingPage } = await import("@/pages/landing/pricing");
    customRender(<PricingPage />);
    await waitFor(() => expect(document.body).toBeInTheDocument());
  });
});

// ─── Auth pages ───────────────────────────────────────────────────────────────

describe("Sign-in page", () => {
  it("renders the login form when user is not authenticated", async () => {
    const { default: SigninPage } = await import("@/pages/auth/signin");
    renderUnauthenticated(<SigninPage />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /login/i })).toBeInTheDocument();
    });
  });

  it("shows 'already logged in' when user is authenticated", async () => {
    const { default: SigninPage } = await import("@/pages/auth/signin");
    customRender(<SigninPage />); // mockAuthContext has user set

    await waitFor(() => {
      expect(screen.getByText(/already logged in/i)).toBeInTheDocument();
    });
  });
});

describe("Sign-up page", () => {
  it("renders the registration form when user is not authenticated", async () => {
    const { default: SignupPage } = await import("@/pages/auth/signup");
    renderUnauthenticated(<SignupPage />);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /create account/i })
      ).toBeInTheDocument();
    });
  });

  it("shows 'already logged in' when user is authenticated", async () => {
    const { default: SignupPage } = await import("@/pages/auth/signup");
    customRender(<SignupPage />); // mockAuthContext has user set

    await waitFor(() => {
      expect(screen.getByText(/already logged in/i)).toBeInTheDocument();
    });
  });
});

describe("Forgot-password page", () => {
  it("renders the form", async () => {
    const { default: ResetPasswordPage } = await import(
      "@/pages/auth/reset-password"
    );
    customRender(<ResetPasswordPage />);
    await waitFor(() => expect(document.body).toBeInTheDocument());
  });
});

// ─── Error / 404 pages ────────────────────────────────────────────────────────

describe("404 page", () => {
  it("renders a not-found message", async () => {
    const { default: NotFoundPage } = await import("@/pages/404");
    customRender(<NotFoundPage />);

    await waitFor(() => {
      expect(screen.getByText(/404/i)).toBeInTheDocument();
    });
  });
});

describe("Not-authorized page", () => {
  it("renders without crashing", async () => {
    const { default: NotAuthorizedPage } = await import(
      "@/pages/not-authorized"
    );
    customRender(<NotAuthorizedPage />);
    await waitFor(() => expect(document.body).toBeInTheDocument());
  });
});

// ─── Authenticated account pages ──────────────────────────────────────────────

describe("Account dashboard", () => {
  it("renders without crashing", async () => {
    const { default: AccountDashboard } = await import(
      "@/pages/account/dashboard"
    );
    customRender(<AccountDashboard />);
    await waitFor(() => expect(document.body).toBeInTheDocument());
  });
});

describe("Account settings (general)", () => {
  it("renders without crashing", async () => {
    const { default: AccountSettingsPage } = await import(
      "@/pages/settings/general-settings"
    );
    customRender(<AccountSettingsPage />);
    await waitFor(() => expect(document.body).toBeInTheDocument());
  });
});

describe("Billing settings", () => {
  it("renders without crashing", async () => {
    const { default: BillingSettingPage } = await import(
      "@/pages/settings/billing-settings"
    );
    customRender(<BillingSettingPage />);
    await waitFor(() => expect(document.body).toBeInTheDocument());
  });
});

// ─── Payment pages ────────────────────────────────────────────────────────────

describe("Payment success page", () => {
  it("renders without crashing", async () => {
    const { default: PaymentSuccessPage } = await import(
      "@/pages/payment/payment-success"
    );
    customRender(<PaymentSuccessPage />);
    await waitFor(() => expect(document.body).toBeInTheDocument());
  });
});

describe("Points payment success page", () => {
  it("renders without crashing", async () => {
    const { default: PointsPaymentSuccessPage } = await import(
      "@/pages/payment/points-payment-success"
    );
    customRender(<PointsPaymentSuccessPage />);
    await waitFor(() => expect(document.body).toBeInTheDocument());
  });
});

// ─── Team pages ───────────────────────────────────────────────────────────────

describe("Team select page", () => {
  it("renders without crashing", async () => {
    const { default: TeamSelect } = await import("@/pages/teams");
    customRender(<TeamSelect />);
    await waitFor(() => expect(document.body).toBeInTheDocument());
  });
});

describe("Account teams page", () => {
  it("renders without crashing", async () => {
    const { default: AccountTeamsPage } = await import(
      "@/pages/account/teams"
    );
    customRender(<AccountTeamsPage />);
    await waitFor(() => expect(document.body).toBeInTheDocument());
  });
});
