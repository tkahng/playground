import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { render } from "@/test/test-utils";
import { mockAuthContext, mockUserTokens } from "@/test/test-utils";
import { AuthContext } from "@/context/auth-context";
import { QueryClientProvider } from "@tanstack/react-query";
import { createTestQueryClient } from "@/test/test-utils";
import React from "react";

vi.mock("@/lib/team-queries", () => ({
  getUserTeamMembers: vi.fn(),
}));

vi.mock("@/components/dashboard-sidebar", () => ({
  DashboardSidebar: () => <div data-testid="sidebar" />,
}));

vi.mock("@/components/create-team-dialog", () => ({
  CreateTeamDialog: ({ trigger }: { trigger?: React.ReactNode }) =>
    trigger ? <>{trigger}</> : <button>Create Team</button>,
}));

vi.mock("@/components/create-team-disabled-tooltip", () => ({
  CreateTeamDisabledTooltip: () => (
    <span data-testid="verify-tooltip">Verify email</span>
  ),
}));

vi.mock("@/components/data-table", () => ({
  DataTable: () => <div data-testid="data-table" />,
}));

import { getUserTeamMembers } from "@/lib/team-queries";
import AccountTeamsPage from "../teams";

function mockTeams(total: number) {
  vi.mocked(getUserTeamMembers).mockResolvedValue({
    data: total > 0 ? [{ team: { name: "Alpha", slug: "alpha" }, role: "owner" }] : [],
    meta: { total, page: 0, per_page: 10 },
  } as any);
}

function renderPage(verified = false) {
  const authValue = verified
    ? {
        ...mockAuthContext,
        user: {
          ...mockUserTokens,
          user: {
            ...mockUserTokens.user,
            email_verified_at: "2024-01-01T00:00:00Z",
          },
        },
      }
    : mockAuthContext;

  return render(
    <QueryClientProvider client={createTestQueryClient()}>
        <AuthContext.Provider value={authValue}>
          <AccountTeamsPage />
        </AuthContext.Provider>
      </QueryClientProvider>,
  );
}

describe("AccountTeamsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("empty state — no teams", () => {
    it("shows 'No teams yet' heading", async () => {
      mockTeams(0);
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("No teams yet")).toBeInTheDocument();
      });
    });

    it("shows descriptive empty-state text", async () => {
      mockTeams(0);
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByText(/create a team to invite collaborators/i),
        ).toBeInTheDocument();
      });
    });

    it("shows disabled create button and verification prompt when not verified", async () => {
      mockTeams(0);
      renderPage(false);
      await waitFor(() => {
        const btn = screen.getByRole("button", { name: /create your first team/i });
        expect(btn).toBeDisabled();
        expect(screen.getByText(/verify your email first/i)).toBeInTheDocument();
      });
    });

    it("shows enabled create button when user is verified", async () => {
      mockTeams(0);
      renderPage(true);
      await waitFor(() => {
        const btn = screen.getByRole("button", { name: /create your first team/i });
        expect(btn).not.toBeDisabled();
      });
    });

    it("does not render the data table", async () => {
      mockTeams(0);
      renderPage();
      await waitFor(() => {
        expect(screen.queryByTestId("data-table")).not.toBeInTheDocument();
      });
    });
  });

  describe("populated state — teams exist", () => {
    it("renders the data table", async () => {
      mockTeams(2);
      renderPage();
      await waitFor(() => {
        expect(screen.getByTestId("data-table")).toBeInTheDocument();
      });
    });

    it("does not show the empty state heading", async () => {
      mockTeams(2);
      renderPage();
      await waitFor(() => {
        expect(screen.queryByText("No teams yet")).not.toBeInTheDocument();
      });
    });
  });
});
