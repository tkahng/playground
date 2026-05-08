import { screen } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { render } from "@/test/test-utils";

vi.mock("@/hooks/use-team", () => ({
  useTeam: vi.fn(() => ({ team: null, setTeam: vi.fn() })),
}));

vi.mock("@/lib/team-queries", () => ({
  createTeam: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import { CreateTeamDialog } from "../create-team-dialog";

function renderDialog(props: React.ComponentProps<typeof CreateTeamDialog> = {}) {
  return render(
    <CreateTeamDialog {...props} />,
  );
}

describe("CreateTeamDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("default trigger", () => {
    it("renders a 'Create Team' button by default", () => {
      renderDialog();
      expect(
        screen.getByRole("button", { name: /create team/i }),
      ).toBeInTheDocument();
    });

    it("default button is disabled when user is not email-verified", () => {
      // mockAuthContext user has no email_verified_at set → disabled
      renderDialog();
      expect(screen.getByRole("button", { name: /create team/i })).toBeDisabled();
    });
  });

  describe("custom trigger prop", () => {
    it("renders the custom trigger element instead of the default button", () => {
      renderDialog({
        trigger: <button>Launch team wizard</button>,
      });
      expect(
        screen.getByRole("button", { name: /launch team wizard/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /^create team$/i }),
      ).not.toBeInTheDocument();
    });

    it("custom trigger can be any valid ReactNode", () => {
      renderDialog({
        trigger: <span role="button">Custom span trigger</span>,
      });
      expect(screen.getByRole("button", { name: /custom span trigger/i })).toBeInTheDocument();
    });
  });
});
