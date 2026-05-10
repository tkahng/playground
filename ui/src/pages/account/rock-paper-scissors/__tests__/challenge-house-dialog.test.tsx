import { render, screen, waitFor } from "@/test/test-utils";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/rps-game-queries", () => ({
  rpsGameQueries: {
    challengeHouse: vi.fn(),
    getLedgerBalance: vi.fn(),
  },
}));

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { rpsGameQueries } from "@/lib/rps-game-queries";
import { toast } from "sonner";
import { ChallengeHouseDialog } from "../challenge-house-dialog";

const mockQueries = rpsGameQueries as unknown as {
  challengeHouse: ReturnType<typeof vi.fn>;
  getLedgerBalance: ReturnType<typeof vi.fn>;
};
const mockToast = toast as unknown as {
  error: ReturnType<typeof vi.fn>;
  success: ReturnType<typeof vi.fn>;
};

// ── helpers ──────────────────────────────────────────────────────────────────

function makeResult(overrides: {
  userMove?: string;
  houseMove?: string;
  userResult?: "win" | "lose" | "tie";
  houseMessage?: string;
  cooldownEndsAt?: string;
} = {}) {
  const {
    userMove = "rock",
    houseMove = "scissors",
    userResult = "win",
    houseMessage,
    cooldownEndsAt = new Date(Date.now() + 5 * 60 * 1000).toISOString(),
  } = overrides;
  return {
    rps_game: {
      id: "game-1",
      status: "completed",
      expires_at: new Date(Date.now() + 30000).toISOString(),
      metadata: {},
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
    requesting_participant: {
      id: "p1",
      player_id: "user-player",
      game_id: "game-1",
      move: userMove,
      result: userResult,
      status: "completed",
      type: "host",
      metadata: {},
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
    invited_participant: {
      id: "p2",
      player_id: "house-id",
      game_id: "game-1",
      move: houseMove,
      result: userResult === "win" ? "lose" : userResult === "lose" ? "win" : "tie",
      status: "completed",
      type: "guest",
      metadata: {},
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
    house_message: houseMessage,
    cooldown_ends_at: cooldownEndsAt,
  };
}

async function openDialog(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("button", { name: /challenge the house/i }));
  await waitFor(() => {
    expect(screen.getByText(/choose your move/i)).toBeInTheDocument();
  });
}

async function selectMoveAndSubmit(
  user: ReturnType<typeof userEvent.setup>,
  move: "Rock" | "Paper" | "Scissors" = "Rock",
) {
  await user.click(screen.getByText(move));
  await user.click(screen.getByRole("button", { name: new RegExp(`play ${move}`, "i") }));
}

// ── tests ────────────────────────────────────────────────────────────────────

describe("ChallengeHouseDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockQueries.getLedgerBalance.mockResolvedValue({ available_balance: 500 });
  });

  // ── initial state ──────────────────────────────────────────────────────────

  describe("initial state", () => {
    it("renders trigger button", () => {
      render(<ChallengeHouseDialog />);
      expect(
        screen.getByRole("button", { name: /challenge the house/i }),
      ).toBeInTheDocument();
    });

    it("shows move selection after opening", async () => {
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);

      expect(screen.getByText("Rock")).toBeInTheDocument();
      expect(screen.getByText("Paper")).toBeInTheDocument();
      expect(screen.getByText("Scissors")).toBeInTheDocument();
    });

    it("submit button is disabled before picking a move", async () => {
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);

      expect(screen.getByRole("button", { name: /select a move/i })).toBeDisabled();
    });

    it("shows house context hint", async () => {
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);

      expect(screen.getByText(/house plays randomly/i)).toBeInTheDocument();
    });

    it("bet toggle is visible", async () => {
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);

      expect(screen.getByLabelText(/add a bet/i)).toBeInTheDocument();
    });
  });

  // ── loading state ──────────────────────────────────────────────────────────

  describe("loading state", () => {
    it("shows thinking indicator while mutation is pending", async () => {
      mockQueries.challengeHouse.mockImplementation(
        () => new Promise(() => {}), // never resolves
      );
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        expect(screen.getByText(/the house is thinking/i)).toBeInTheDocument();
      });
    });

    it("hides move selection while pending", async () => {
      mockQueries.challengeHouse.mockImplementation(() => new Promise(() => {}));
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        expect(screen.queryByText(/choose your move/i)).not.toBeInTheDocument();
      });
    });
  });

  // ── result — user wins ─────────────────────────────────────────────────────

  describe("result: user wins", () => {
    it("shows Victory on win", async () => {
      mockQueries.challengeHouse.mockResolvedValue(makeResult({ userResult: "win" }));
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        expect(screen.getByText("Victory!")).toBeInTheDocument();
      });
    });

    it("shows the house as opponent", async () => {
      mockQueries.challengeHouse.mockResolvedValue(makeResult({ userResult: "win" }));
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        // "Played against" label only appears in the GameResult opponent section
        expect(screen.getByText(/played against/i)).toBeInTheDocument();
      });
    });

    it("shows both moves", async () => {
      mockQueries.challengeHouse.mockResolvedValue(
        makeResult({ userMove: "rock", houseMove: "scissors", userResult: "win" }),
      );
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user, "Rock");

      await waitFor(() => {
        expect(screen.getByText("rock")).toBeInTheDocument();
        expect(screen.getByText("scissors")).toBeInTheDocument();
      });
    });

    it("does NOT show house_message when user wins", async () => {
      mockQueries.challengeHouse.mockResolvedValue(
        makeResult({ userResult: "win", houseMessage: undefined }),
      );
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        expect(screen.queryByText(/house always wins/i)).not.toBeInTheDocument();
      });
    });

    it("shows cooldown end time after result", async () => {
      const cooldown = new Date("2026-05-10T20:00:00Z").toISOString();
      mockQueries.challengeHouse.mockResolvedValue(
        makeResult({ userResult: "win", cooldownEndsAt: cooldown }),
      );
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        expect(screen.getByText(/next challenge available after/i)).toBeInTheDocument();
      });
    });

    it("shows play-again-later button", async () => {
      mockQueries.challengeHouse.mockResolvedValue(makeResult({ userResult: "win" }));
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: /play again later/i }),
        ).toBeInTheDocument();
      });
    });
  });

  // ── result — house wins ────────────────────────────────────────────────────

  describe("result: house wins", () => {
    it("shows Defeat on loss", async () => {
      mockQueries.challengeHouse.mockResolvedValue(makeResult({ userResult: "lose" }));
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        expect(screen.getByText("Defeat")).toBeInTheDocument();
      });
    });

    it("shows house_message catchphrase when present", async () => {
      mockQueries.challengeHouse.mockResolvedValue(
        makeResult({ userResult: "lose", houseMessage: "House always wins." }),
      );
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        expect(screen.getByText("House always wins.")).toBeInTheDocument();
      });
    });

    it("does NOT show house_message when absent (no bet)", async () => {
      mockQueries.challengeHouse.mockResolvedValue(
        makeResult({ userResult: "lose", houseMessage: undefined }),
      );
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        expect(screen.queryByText(/house always wins/i)).not.toBeInTheDocument();
      });
    });
  });

  // ── result — tie ──────────────────────────────────────────────────────────

  describe("result: tie", () => {
    it("shows It's a Tie on tie", async () => {
      mockQueries.challengeHouse.mockResolvedValue(makeResult({ userResult: "tie" }));
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        expect(screen.getByText(/it's a tie/i)).toBeInTheDocument();
      });
    });
  });

  // ── reset ─────────────────────────────────────────────────────────────────

  describe("reset", () => {
    it("returns to move selection after clicking play-again-later", async () => {
      mockQueries.challengeHouse.mockResolvedValue(makeResult({ userResult: "win" }));
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => screen.getByRole("button", { name: /play again later/i }));
      await user.click(screen.getByRole("button", { name: /play again later/i }));

      await waitFor(() => {
        expect(screen.getByText(/choose your move/i)).toBeInTheDocument();
      });
    });
  });

  // ── error handling ────────────────────────────────────────────────────────

  describe("error handling", () => {
    it("shows error toast when API call fails", async () => {
      mockQueries.challengeHouse.mockRejectedValue(new Error("cooldown active"));
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalledWith("cooldown active");
      });
    });

    it("returns to move selection on error", async () => {
      mockQueries.challengeHouse.mockRejectedValue(new Error("cooldown active"));
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user);

      await waitFor(() => {
        expect(screen.getByText(/choose your move/i)).toBeInTheDocument();
      });
    });
  });

  // ── bet section ───────────────────────────────────────────────────────────

  describe("bet section", () => {
    it("bet input hidden by default", async () => {
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);

      expect(
        screen.queryByPlaceholderText(/1 – 500 pts/i),
      ).not.toBeInTheDocument();
    });

    it("bet input appears after toggling on", async () => {
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await user.click(screen.getByLabelText(/add a bet/i));

      await waitFor(() => {
        expect(screen.getByPlaceholderText(/1 – 500 pts/i)).toBeInTheDocument();
      });
    });

    it("shows available balance when bet enabled", async () => {
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await user.click(screen.getByLabelText(/add a bet/i));

      await waitFor(() => {
        // "Available:" prefix is unique to the balance display paragraph
        expect(screen.getByText(/^Available:/)).toBeInTheDocument();
      });
    });

    it("input max is capped at 500 when balance exceeds max", async () => {
      mockQueries.getLedgerBalance.mockResolvedValue({ available_balance: 1000 });
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await user.click(screen.getByLabelText(/add a bet/i));

      await waitFor(() => {
        const input = screen.getByPlaceholderText(/1 – 500 pts/i);
        expect(input).toHaveAttribute("max", "500");
      });
    });

    it("input max is capped at balance when balance is below 500", async () => {
      mockQueries.getLedgerBalance.mockResolvedValue({ available_balance: 120 });
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await user.click(screen.getByLabelText(/add a bet/i));

      await waitFor(() => {
        const input = screen.getByPlaceholderText(/1 – 500 pts/i);
        expect(input).toHaveAttribute("max", "120");
      });
    });

    it("shows buy-points link when balance is 0", async () => {
      mockQueries.getLedgerBalance.mockResolvedValue({ available_balance: 0 });
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await user.click(screen.getByLabelText(/add a bet/i));

      await waitFor(() => {
        expect(screen.getByText(/buy points/i)).toBeInTheDocument();
        expect(screen.queryByPlaceholderText(/1 – 500 pts/i)).not.toBeInTheDocument();
      });
    });

    it("disabling bet toggle hides input", async () => {
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);

      const toggle = screen.getByLabelText(/add a bet/i);
      await user.click(toggle);
      await waitFor(() =>
        expect(screen.getByPlaceholderText(/1 – 500 pts/i)).toBeInTheDocument(),
      );

      await user.click(toggle);
      await waitFor(() =>
        expect(screen.queryByPlaceholderText(/1 – 500 pts/i)).not.toBeInTheDocument(),
      );
    });

    it("passes bet_amount to challengeHouse when bet is set", async () => {
      mockQueries.challengeHouse.mockResolvedValue(makeResult({ userResult: "win" }));
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);

      await user.click(screen.getByLabelText(/add a bet/i));
      await waitFor(() => screen.getByPlaceholderText(/1 – 500 pts/i));
      await user.type(screen.getByPlaceholderText(/1 – 500 pts/i), "100");

      await selectMoveAndSubmit(user, "Rock");

      await waitFor(() => {
        expect(mockQueries.challengeHouse).toHaveBeenCalledWith(
          expect.objectContaining({ betAmount: 100 }),
        );
      });
    });

    it("does not pass bet_amount when bet toggle is off", async () => {
      mockQueries.challengeHouse.mockResolvedValue(makeResult({ userResult: "win" }));
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user, "Rock");

      await waitFor(() => {
        expect(mockQueries.challengeHouse).toHaveBeenCalledWith(
          expect.objectContaining({ betAmount: undefined }),
        );
      });
    });

    it("shows bet outcome in result when bet was placed", async () => {
      mockQueries.challengeHouse.mockResolvedValue(
        makeResult({ userResult: "win", houseMessage: undefined }),
      );
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);

      await user.click(screen.getByLabelText(/add a bet/i));
      await waitFor(() => screen.getByPlaceholderText(/1 – 500 pts/i));
      await user.type(screen.getByPlaceholderText(/1 – 500 pts/i), "50");
      await selectMoveAndSubmit(user, "Rock");

      await waitFor(() => {
        expect(screen.getByText("+50 pts")).toBeInTheDocument();
      });
    });
  });

  // ── mutation args ─────────────────────────────────────────────────────────

  describe("mutation call", () => {
    it("calls challengeHouse with correct move", async () => {
      mockQueries.challengeHouse.mockResolvedValue(makeResult({ userResult: "win" }));
      const user = userEvent.setup();
      render(<ChallengeHouseDialog />);
      await openDialog(user);
      await selectMoveAndSubmit(user, "Scissors");

      await waitFor(() => {
        expect(mockQueries.challengeHouse).toHaveBeenCalledWith(
          expect.objectContaining({ move: "scissors" }),
        );
      });
    });
  });
});
