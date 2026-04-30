import { render, screen, waitFor } from "@/test/test-utils";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@/lib/rps-game-queries", () => ({
  rpsGameQueries: {
    findPlayerByEmail: vi.fn(),
    getLedgerBalance: vi.fn(),
    requestGame: vi.fn(),
    requestGameEmail: vi.fn(),
  },
}));

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { rpsGameQueries } from "@/lib/rps-game-queries";
import { CreateGameDialog } from "../create-game-dialog";

const mockQueries = rpsGameQueries as unknown as {
  findPlayerByEmail: ReturnType<typeof vi.fn>;
  getLedgerBalance: ReturnType<typeof vi.fn>;
  requestGame: ReturnType<typeof vi.fn>;
  requestGameEmail: ReturnType<typeof vi.fn>;
};

const fakePlayer = {
  id: "player-99",
  email: "opponent@example.com",
  display_name: "Opponent",
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

// Interacts with the dialog to search for and find a player.
// Mocks must be configured BEFORE calling render() — TanStack Query fires the
// balance query on mount, before this helper executes.
async function openDialogAndFindPlayer(
  user: ReturnType<typeof userEvent.setup>,
) {
  mockQueries.findPlayerByEmail.mockResolvedValue({ data: fakePlayer });

  await user.click(screen.getByRole("button", { name: /play a game/i }));
  await user.type(
    screen.getByPlaceholderText(/enter full email address/i),
    "opponent@example.com",
  );
  await user.click(screen.getByRole("button", { name: /search/i }));

  await waitFor(() => {
    expect(screen.getByLabelText(/add a bet/i)).toBeInTheDocument();
  });
}

describe("CreateGameDialog — bet section", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockQueries.requestGame.mockResolvedValue({ data: {} });
    mockQueries.requestGameEmail.mockResolvedValue({ data: {} });
  });

  it("bet toggle not visible before player is found", async () => {
    mockQueries.getLedgerBalance.mockResolvedValue({ available_balance: 100 });
    const user = userEvent.setup();
    render(<CreateGameDialog />);

    await user.click(screen.getByRole("button", { name: /play a game/i }));

    expect(screen.queryByLabelText(/add a bet/i)).not.toBeInTheDocument();
  });

  it("bet toggle appears after player found", async () => {
    mockQueries.getLedgerBalance.mockResolvedValue({ available_balance: 200 });
    const user = userEvent.setup();
    render(<CreateGameDialog />);

    await openDialogAndFindPlayer(user);

    expect(screen.getByLabelText(/add a bet/i)).toBeInTheDocument();
  });

  it("amount input hidden when bet toggle is off", async () => {
    mockQueries.getLedgerBalance.mockResolvedValue({ available_balance: 200 });
    const user = userEvent.setup();
    render(<CreateGameDialog />);

    await openDialogAndFindPlayer(user);

    expect(
      screen.queryByPlaceholderText(/enter bet amount/i),
    ).not.toBeInTheDocument();
  });

  it("amount input appears after enabling bet toggle", async () => {
    mockQueries.getLedgerBalance.mockResolvedValue({ available_balance: 200 });
    const user = userEvent.setup();
    render(<CreateGameDialog />);

    await openDialogAndFindPlayer(user);
    await user.click(screen.getByLabelText(/add a bet/i));

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText(/enter bet amount/i),
      ).toBeInTheDocument();
    });
  });

  it("shows available balance when bet enabled and balance > 0", async () => {
    mockQueries.getLedgerBalance.mockResolvedValue({ available_balance: 150 });
    const user = userEvent.setup();
    render(<CreateGameDialog />);

    await openDialogAndFindPlayer(user);
    await user.click(screen.getByLabelText(/add a bet/i));

    await waitFor(() => {
      expect(screen.getByText(/150 pts/)).toBeInTheDocument();
    });
  });

  it("shows buy-points link when balance is 0", async () => {
    mockQueries.getLedgerBalance.mockResolvedValue({ available_balance: 0 });
    const user = userEvent.setup();
    render(<CreateGameDialog />);

    await openDialogAndFindPlayer(user);
    await user.click(screen.getByLabelText(/add a bet/i));

    await waitFor(() => {
      expect(
        screen.getByRole("link", { name: /buy points/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByPlaceholderText(/enter bet amount/i),
      ).not.toBeInTheDocument();
    });
  });

  it("amount input max is capped at available balance", async () => {
    mockQueries.getLedgerBalance.mockResolvedValue({ available_balance: 75 });
    const user = userEvent.setup();
    render(<CreateGameDialog />);

    await openDialogAndFindPlayer(user);
    await user.click(screen.getByLabelText(/add a bet/i));

    await waitFor(() => {
      const input = screen.getByPlaceholderText(/enter bet amount/i);
      expect(input).toHaveAttribute("max", "75");
    });
  });

  it("disabling bet toggle hides amount input", async () => {
    mockQueries.getLedgerBalance.mockResolvedValue({ available_balance: 100 });
    const user = userEvent.setup();
    render(<CreateGameDialog />);

    await openDialogAndFindPlayer(user);

    const toggle = screen.getByLabelText(/add a bet/i);
    await user.click(toggle);

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText(/enter bet amount/i),
      ).toBeInTheDocument();
    });

    await user.click(toggle);

    await waitFor(() => {
      expect(
        screen.queryByPlaceholderText(/enter bet amount/i),
      ).not.toBeInTheDocument();
    });
  });
});
