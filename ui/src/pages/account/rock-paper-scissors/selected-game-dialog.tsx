import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { rpsGameQueries } from "@/lib/rps-game-queries";
import { Participant, PlayerRpsGame } from "@/schema.types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { Move, MoveSelection } from "./move";
import { GameResult } from "./game-result";

export const SelectedRpsGameDialog = ({
  dialogProps,
  onClose,
  gameId,
}: {
  dialogProps: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
  };
  onClose: () => void;
  gameId: string | null;
}) => {
  const userInfo = useAuthProvider();
  const {
    data: selectedGame,
    isLoading: selectedRpsGameLoading,
    isError: isSelectedRpsGameError,
    error: selectedRpsGameError,
  } = useQuery({
    enabled: !!gameId,
    queryKey: [{ key: "find-rps-game", id: gameId }],
    queryFn: async () => {
      if (!userInfo.user?.tokens.access_token) {
        throw new Error("No access token");
      }
      if (!gameId) {
        throw new Error("No game ID");
      }
      const { data: game } = await rpsGameQueries.getRpsGame({
        token: userInfo.user.tokens.access_token,
        gameId: gameId,
      });
      if (!game) {
        return {
          data: null,
        };
      }
      let player: Participant;
      let opponent: Participant;
      if (game.invited_participant.player_id === userInfo.user.user.id) {
        player = game.invited_participant;
        opponent = game.requesting_participant;
      } else {
        player = game.requesting_participant;
        opponent = game.invited_participant;
      }
      return {
        data: {
          rpsGame: game.rps_game,
          player,
          opponent,
        },
      };
    },
  });
  if (selectedRpsGameLoading) {
    return (
      <Dialog {...dialogProps}>
        <DialogContent
          onCloseAutoFocus={() => {
            onClose();
          }}
        >
          <DialogTitle>Loading...</DialogTitle>
        </DialogContent>
      </Dialog>
    );
  }
  if (isSelectedRpsGameError) {
    return (
      <Dialog {...dialogProps}>
        <DialogContent
          onCloseAutoFocus={() => {
            onClose();
          }}
        >
          <DialogTitle>{selectedRpsGameError.message}</DialogTitle>
        </DialogContent>
      </Dialog>
    );
  }
  const expired =
    new Date(selectedGame?.data?.rpsGame.expires_at || 0).getTime() <
      // eslint-disable-next-line react-hooks/purity
      Date.now() && selectedGame?.data?.rpsGame.status === "pending";
  return (
    <Dialog {...dialogProps}>
      <DialogContent
        onCloseAutoFocus={() => {
          onClose();
        }}
      >
        <DialogTitle>
          Your game against{" "}
          {selectedGame?.data?.opponent.player?.display_name ||
            selectedGame?.data?.opponent.player?.email}
        </DialogTitle>

        <div className="flex flex-col gap-4">
          {/* no game found */}
          {!selectedGame?.data && (
            <NoPlayerView onOpenChange={dialogProps.onOpenChange} />
          )}

          {/* game found with pending status and you have submitted your move */}
          {selectedGame?.data &&
            !expired &&
            selectedGame.data.rpsGame.status === "pending" &&
            selectedGame.data.player.status === "completed" && (
              <PendingGameView
                onOpenChange={dialogProps.onOpenChange}
                game={selectedGame.data}
              />
            )}

          {/* game found with pending status and you have not submitted your move */}
          {selectedGame?.data &&
            !expired &&
            selectedGame.data.rpsGame.status === "pending" &&
            selectedGame.data.player.status === "pending" && (
              <SubmitMoveView
                onOpenChange={dialogProps.onOpenChange}
                game={selectedGame.data}
              />
            )}
          {selectedGame?.data &&
            !expired &&
            selectedGame.data.rpsGame.status === "completed" && (
              <GameResult
                result={selectedGame.data.player.result}
                opponent={selectedGame.data.opponent.player?.email || ""}
                playerMove={selectedGame.data.player.move}
                opponentMove={selectedGame.data.opponent.move}
                betAmount={selectedGame.data.rpsGame.bet_amount}
                betResult={selectedGame.data.player.result}
              />
            )}
          {selectedGame && expired && (
            <ExpiredGameView onOpenChange={dialogProps.onOpenChange} />
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};

export const ExpiredGameView = ({
  onOpenChange,
}: {
  onOpenChange: (open: boolean) => void;
}) => {
  return (
    <div>
      <p className="font-bold text-lg">This game is expired.</p>
      <Button onClick={() => onOpenChange(false)}>Close</Button>
    </div>
  );
};

export const NoPlayerView = ({
  onOpenChange,
}: {
  onOpenChange: (open: boolean) => void;
}) => {
  return (
    <div>
      <p className="font-bold text-lg">No game selected.</p>
      <Button onClick={() => onOpenChange(false)}>Close</Button>
    </div>
  );
};

function useExpiryWarning(expiresAt: string): string | null {
  const ms = new Date(expiresAt).getTime() - Date.now();
  if (ms <= 0) return null;
  const hours = ms / (1000 * 60 * 60);
  if (hours > 24) return null;
  if (hours < 1) return `less than 1 hour`;
  return `${Math.floor(hours)} hour${Math.floor(hours) === 1 ? "" : "s"}`;
}

export const PendingGameView = ({
  onOpenChange,
  game,
}: {
  onOpenChange: (open: boolean) => void;
  game: PlayerRpsGame;
}) => {
  const opponentName =
    game.opponent.player?.display_name || game.opponent.player?.email || "your opponent";
  const betAmount = game.rpsGame.bet_amount;
  const expiryWarning = useExpiryWarning(game.rpsGame.expires_at);

  return (
    <div className="flex flex-col items-center gap-4 py-4 text-center">
      <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-muted">
        <span className="text-3xl">⏳</span>
      </div>
      <div>
        <p className="text-lg font-semibold">Waiting for {opponentName}</p>
        <p className="text-sm text-muted-foreground mt-1">
          Your move is locked in. The result will appear once they respond.
        </p>
      </div>
      {expiryWarning && (
        <div className="rounded-lg border border-orange-300 bg-orange-50 px-4 py-2 text-sm text-orange-800">
          ⚠️ This game expires in {expiryWarning}
        </div>
      )}
      {betAmount != null && betAmount > 0 && (
        <div className="rounded-lg border border-amber-300 bg-amber-50 px-4 py-2 text-sm text-amber-800">
          🪙 {betAmount} pts on the line
        </div>
      )}
      <Button variant="outline" onClick={() => onOpenChange(false)}>
        Close
      </Button>
    </div>
  );
};

export const SubmitMoveView = ({
  game,
}: {
  onOpenChange: (open: boolean) => void;
  game: PlayerRpsGame;
}) => {
  const { user } = useAuthProvider();
  const queryClient = useQueryClient();

  const balanceQuery = useQuery({
    queryKey: [{ key: "ledger-balance" }],
    queryFn: () =>
      rpsGameQueries.getLedgerBalance({ token: user!.tokens.access_token }),
    enabled: !!user,
  });

  const betAmount = game.rpsGame.bet_amount ?? 0;
  const availableBalance = balanceQuery.data?.available_balance ?? 0;
  const insufficientFunds = betAmount > 0 && availableBalance < betAmount;

  const submitToGameMutation = useMutation({
    mutationFn: async (data: { move: Move }) => {
      if (!user) {
        throw new Error("No user");
      }

      return rpsGameQueries.submitMoveToGame({
        token: user?.tokens.access_token,
        move: data.move,
        gameId: game.rpsGame.id,
        status: "completed",
      });
    },
    onError: (error) => {
      toast.error(error.message);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: [{ key: "find-rps-game" }],
      });
      await queryClient.invalidateQueries({
        queryKey: [{ key: "rps-games" }],
      });
      await queryClient.invalidateQueries({
        queryKey: [{ key: "ledger-balance" }],
      });
      toast.success("move submitted");
    },
  });

  return (
    <div>
      <MoveSelection
        opponentPlayer={game.opponent.player}
        handleSubmit={(move) => submitToGameMutation.mutate({ move })}
        disabled={insufficientFunds}
      >
        {insufficientFunds && (
          <p className="text-center text-sm text-destructive mb-2">
            You need {betAmount} pts to accept this bet but only have{" "}
            {availableBalance} pts.{" "}
            <a href="/account/settings/points" className="underline">
              Buy points
            </a>
          </p>
        )}
      </MoveSelection>
    </div>
  );
};
