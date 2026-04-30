import { render, screen, waitFor } from "@/test/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/rps-game-queries", () => ({
  rpsGameQueries: {
    getLedgerBalance: vi.fn(),
    submitMoveToGame: vi.fn(),
  },
}));

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { rpsGameQueries } from "@/lib/rps-game-queries";
import type { PlayerRpsGame } from "@/schema.types";
import { SubmitMoveView } from "../selected-game-dialog";

const mockLedger = rpsGameQueries as {
  getLedgerBalance: ReturnType<typeof vi.fn>;
  submitMoveToGame: ReturnType<typeof vi.fn>;
};

function makeGame(betAmount: number): PlayerRpsGame {
  return {
    rpsGame: {
      id: "game-1",
      status: "pending",
      bet_amount: betAmount,
      expires_at: new Date(Date.now() + 3_600_000).toISOString(),
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-01-01T00:00:00Z",
      metadata: "{}",
    },
    player: {
      id: "part-1",
      game_id: "game-1",
      player_id: "user-1",
      move: "rock",
      status: "pending",
      result: "tie",
      type: "host",
      created_at: "2024-01-01T00:00:00Z",
      metadata: "{}",
    },
    opponent: {
      id: "part-2",
      game_id: "game-1",
      player_id: "user-2",
      player: {
        id: "user-2",
        email: "opponent@example.com",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      } as PlayerRpsGame["opponent"]["player"],
      move: "scissors",
      status: "pending",
      result: "tie",
      type: "guest",
      created_at: "2024-01-01T00:00:00Z",
      metadata: "{}",
    },
  };
}

describe("SubmitMoveView — insufficient funds", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows all moves enabled when balance covers bet", async () => {
    mockLedger.getLedgerBalance.mockResolvedValue({ available_balance: 500 });
    render(
      <SubmitMoveView
        game={makeGame(100)}
        onOpenChange={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(
        screen.queryByText(/you need.*pts to accept this bet/i)
      ).not.toBeInTheDocument();
    });

    expect(screen.getByRole("button", { name: "Select a Move" })).toBeDisabled();
    // disabled only because no move selected, not because of funds
    expect(
      screen.queryByText(/buy points/i)
    ).not.toBeInTheDocument();
  });

  it("shows insufficient funds warning when balance < bet", async () => {
    mockLedger.getLedgerBalance.mockResolvedValue({ available_balance: 10 });
    render(
      <SubmitMoveView
        game={makeGame(100)}
        onOpenChange={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(
        screen.getByText(/you need 100 pts to accept this bet but only have 10 pts/i)
      ).toBeInTheDocument();
    });

    expect(screen.getByRole("link", { name: /buy points/i })).toBeInTheDocument();
  });

  it("submit button disabled due to insufficient funds even after selecting a move", async () => {
    mockLedger.getLedgerBalance.mockResolvedValue({ available_balance: 5 });
    const { userEvent: user } = await import("@testing-library/user-event");
    const ue = user.setup();

    render(
      <SubmitMoveView
        game={makeGame(50)}
        onOpenChange={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText(/you need 50 pts/i)).toBeInTheDocument();
    });

    await ue.click(screen.getByText("Rock"));
    expect(screen.getByRole("button", { name: "Play Rock" })).toBeDisabled();
  });

  it("no warning shown when game has no bet", async () => {
    mockLedger.getLedgerBalance.mockResolvedValue({ available_balance: 0 });
    render(
      <SubmitMoveView
        game={makeGame(0)}
        onOpenChange={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(
        screen.queryByText(/pts to accept this bet/i)
      ).not.toBeInTheDocument();
    });
  });
});
