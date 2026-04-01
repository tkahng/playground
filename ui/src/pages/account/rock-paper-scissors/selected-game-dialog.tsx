import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
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
      if (game.invited_participant.player?.email === userInfo.user.user.email) {
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

export const PendingGameView = ({
  game,
}: {
  onOpenChange: (open: boolean) => void;
  game: PlayerRpsGame;
}) => {
  const parti = [game.player, game.opponent];
  return (
    <div>
      <p className="font-bold text-lg">
        Result:{" "}
        {game.rpsGame.status === "completed"
          ? game.player.result
          : game.rpsGame.status}
      </p>
      <div className="flex flex-row gap-4">
        {parti.map((m) => (
          <Card key={m.id} className="grow flex">
            Move: {m.move}
          </Card>
        ))}
      </div>
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
      toast.success("move submitted");
    },
  });

  return (
    <div>
      <MoveSelection
        opponentPlayer={game.opponent.player}
        handleSubmit={(move) => submitToGameMutation.mutate({ move })}
      />
    </div>
  );
};
