import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { render } from "@/test/test-utils";
import type { OnboardingProgress } from "@/hooks/use-onboarding-progress";

vi.mock("@/hooks/use-onboarding-progress");
vi.mock("@/lib/team-queries");
vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
  Toaster: () => null,
}));

import { useOnboardingProgress } from "@/hooks/use-onboarding-progress";
import { getUserTeamMembers } from "@/lib/team-queries";
import { OnboardingChecklist } from "../onboarding-checklist";

const mockDismiss = vi.fn();

const ALL_INCOMPLETE: OnboardingProgress = {
  saidHello: false,
  hasProject: false,
  visitedPricing: false,
  visitedRps: false,
  dismissed: false,
};

function setupMocks(
  progressOverrides: Partial<OnboardingProgress> = {},
  teamTotal = 0,
) {
  vi.mocked(useOnboardingProgress).mockReturnValue({
    progress: { ...ALL_INCOMPLETE, ...progressOverrides },
    markStep: vi.fn(),
    dismiss: mockDismiss,
  });
  vi.mocked(getUserTeamMembers).mockResolvedValue({
    data: [],
    meta: { total: teamTotal, page: 0, per_page: 1 },
  } as any);
}

function renderChecklist() {
  return render(
    <MemoryRouter>
      <OnboardingChecklist />
    </MemoryRouter>,
  );
}

describe("OnboardingChecklist", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the progress heading", async () => {
    setupMocks();
    renderChecklist();
    await waitFor(() => {
      expect(screen.getByText(/get started/i)).toBeInTheDocument();
    });
  });

  it("shows 0/5 when no steps are complete", async () => {
    setupMocks();
    renderChecklist();
    await waitFor(() => {
      // The counter is rendered as "{completed}/{total}" in a <span>; use the
      // parent CardTitle which has normalised text content "Get started — 0 / 5 steps done"
      const title = screen.getByText(/steps done/i);
      expect(title).toHaveTextContent("0");
      expect(title).toHaveTextContent("5");
    });
  });

  it("renders all 5 step labels", async () => {
    setupMocks();
    renderChecklist();
    await waitFor(() => {
      expect(screen.getByText("Say Hello to the world")).toBeInTheDocument();
      expect(screen.getByText("Create your first team")).toBeInTheDocument();
      expect(screen.getByText("Create a project")).toBeInTheDocument();
      expect(screen.getByText("Check out plans")).toBeInTheDocument();
      expect(screen.getByText("Play Rock Paper Scissors")).toBeInTheDocument();
    });
  });

  it("shows checked icon for completed localStorage steps", async () => {
    setupMocks({ saidHello: true, visitedPricing: true });
    renderChecklist();
    await waitFor(() => {
      // Completed steps show strikethrough text
      const saidHelloLabel = screen.getByText("Say Hello to the world");
      expect(saidHelloLabel).toHaveClass("line-through");
      const pricingLabel = screen.getByText("Check out plans");
      expect(pricingLabel).toHaveClass("line-through");
    });
  });

  it("hasTeam step is checked when API returns teams", async () => {
    setupMocks({}, 1);
    renderChecklist();
    await waitFor(() => {
      const teamLabel = screen.getByText("Create your first team");
      expect(teamLabel).toHaveClass("line-through");
    });
  });

  it("hasTeam step is unchecked when API returns no teams", async () => {
    setupMocks({}, 0);
    renderChecklist();
    await waitFor(() => {
      const teamLabel = screen.getByText("Create your first team");
      expect(teamLabel).not.toHaveClass("line-through");
    });
  });

  it("returns null when progress.dismissed is true", () => {
    setupMocks({ dismissed: true });
    const { container } = renderChecklist();
    expect(container).toBeEmptyDOMElement();
  });

  it("dismiss button calls the dismiss handler", async () => {
    setupMocks();
    renderChecklist();
    const user = userEvent.setup();
    await waitFor(() => {
      expect(screen.getByTitle("Dismiss")).toBeInTheDocument();
    });
    await user.click(screen.getByTitle("Dismiss"));
    expect(mockDismiss).toHaveBeenCalledOnce();
  });

  it("shows all-done message when all 5 steps complete", async () => {
    setupMocks(
      {
        saidHello: true,
        hasProject: true,
        visitedPricing: true,
        visitedRps: true,
      },
      1, // hasTeam via API
    );
    renderChecklist();
    await waitFor(() => {
      expect(
        screen.getByText(/you've explored everything/i),
      ).toBeInTheDocument();
    });
  });

  it("renders CTA links for incomplete steps", async () => {
    setupMocks();
    renderChecklist();
    await waitFor(() => {
      expect(screen.getByRole("link", { name: /say hello/i })).toBeInTheDocument();
      expect(screen.getByRole("link", { name: /see plans/i })).toBeInTheDocument();
      expect(screen.getByRole("link", { name: /play now/i })).toBeInTheDocument();
    });
  });

  it("hides CTA link for a completed step", async () => {
    setupMocks({ visitedPricing: true });
    renderChecklist();
    await waitFor(() => {
      expect(screen.queryByRole("link", { name: /see plans/i })).not.toBeInTheDocument();
    });
  });
});
