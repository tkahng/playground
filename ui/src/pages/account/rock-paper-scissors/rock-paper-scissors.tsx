import { useSearchParams } from "@/hooks/use-search-params";
import { DataTable } from "@/components/data-table";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { rpsGameQueries } from "@/lib/rps-game-queries";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { Link } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { SelectedRpsGameDialog } from "./selected-game-dialog";
import { Participant, PlayerRpsGame } from "@/schema.types";
import { ClassValue } from "clsx";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { CreateGameDialog } from "./create-game-dialog";
import { CenteredSpinner } from "@/components/centered-spinner";
import { RouteMap } from "@/components/route-map";

export default function RockPaperScissors() {
  const userInfo = useAuthProvider();
  const [searchParams, setSearchParams] = useSearchParams();
  const gameId = searchParams.get("game_id");

  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);
  const onClickGameId = (key: string | null) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev.toString());
      if (key) {
        next.set("game_id", key);
      } else {
        next.delete("game_id");
      }
      return next;
    });
  };
  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const newState =
      typeof updater === "function"
        ? updater({ pageIndex, pageSize })
        : updater;
    setSearchParams({
      page: String(newState.pageIndex),
      per_page: String(newState.pageSize),
    });
  };
  const { data, isLoading, isError, error } = useQuery({
    queryKey: [{ key: "rps-games", page: pageIndex, per_page: pageSize }],
    queryFn: async () => {
      if (!userInfo.user?.tokens.access_token) {
        throw new Error("No access token");
      }
      const games = await rpsGameQueries.getRpsGames({
        token: userInfo.user.tokens.access_token,
        page: pageIndex,
        per_page: pageSize,
        sort_order: "desc",
        sort_by: "created_at",
      });
      const playerGames: PlayerRpsGame[] = [];
      if (games.data?.length) {
        for (const game of games.data) {
          let player: Participant;
          let opponent: Participant;
          if (
            game.invited_participant.player?.email === userInfo.user.user.email
          ) {
            player = game.invited_participant;
            opponent = game.requesting_participant;
          } else {
            player = game.requesting_participant;
            opponent = game.invited_participant;
          }
          playerGames.push({
            rpsGame: game.rps_game,
            player,
            opponent,
          });
        }
      }
      return {
        data: playerGames,
        meta: games.meta,
      };
    },
  });

  const { data: balanceData } = useQuery({
    queryKey: [{ key: "ledger-balance" }],
    enabled: !!userInfo.user?.tokens.access_token,
    queryFn: async () => {
      if (!userInfo.user?.tokens.access_token) {
        throw new Error("No access token");
      }
      return rpsGameQueries.getLedgerBalance({
        token: userInfo.user.tokens.access_token,
      });
    },
  });

  const [open, onOpenChange] = useState(!!gameId);
  useEffect(() => {
    if (gameId) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      onOpenChange(true);
    } else {
      onOpenChange(false);
    }
  }, [gameId]);
  if (isLoading) {
    return <CenteredSpinner />;
  }
  if (isError) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div>
      <div className="flex items-center gap-3 mb-1">
        <h1>Rock Paper Scissors</h1>
        {balanceData && (
          balanceData.available_balance <= 0 ? (
            <Link
              to={RouteMap.POINTS_SETTINGS}
              className="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-sm font-medium text-amber-600 underline underline-offset-2 hover:text-amber-700"
            >
              🪙 0 pts — Buy Points
            </Link>
          ) : (
            <span className="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-sm font-medium">
              🪙 {balanceData.available_balance ?? 0} pts
            </span>
          )
        )}
      </div>
      <div className="flex items-center justify-between">
        <p>Start a new Game with a friend</p>
        <CreateGameDialog />
      </div>
      <DataTable
        columns={[
          {
            header: "Result",
            cell: ({ row }) => {
              const state = CalculateGameState(row.original);
              const style = GameStateClassValMap[state];
              return <Badge className={cn(style)}>{state}</Badge>;
            },
          },
          {
            header: "Your Move",
            cell: ({ row }) => {
              if (row.original.player?.status === "completed") {
                return row.original.player.move;
              }
              if (row.original.player?.status === "pending") {
                return "Pending";
              }
              if (row.original.player?.status === "declined") {
                return "Declined";
              }
              return "";
            },
          },
          {
            header: "Opponent",
            cell: ({ row }) => {
              return row.original.opponent?.player?.email || "";
            },
          },
          {
            header: "Created At",
            cell: ({ row }) => {
              return new Date(
                row.original.rpsGame.created_at,
              ).toLocaleDateString();
            },
          },
          {
            header: "Bet",
            cell: ({ row }) => {
              const betAmount = row.original.rpsGame.bet_amount;
              if (!betAmount) {
                return <span className="text-muted-foreground">—</span>;
              }
              const state = CalculateGameState(row.original);
              if (state === GameState.Win) {
                return (
                  <span className="text-green-600 font-semibold">
                    +{betAmount} pts
                  </span>
                );
              }
              if (state === GameState.Lose) {
                return (
                  <span className="text-red-600 font-semibold">
                    −{betAmount} pts
                  </span>
                );
              }
              if (
                state === GameState.Tie ||
                state === GameState.Cancelled ||
                state === GameState.Expired
              ) {
                return <span className="text-muted-foreground">refunded</span>;
              }
              // Pending or Submitted
              return (
                <span className="text-amber-600">{betAmount} pts at stake</span>
              );
            },
          },
        ]}
        onClick={(row) => {
          onClickGameId(row.original.rpsGame.id);
        }}
        data={data?.data || []}
        rowCount={data?.meta.total || 0}
        paginationState={{ pageIndex, pageSize }}
        onPaginationChange={onPaginationChange}
        paginationEnabled
      />
      <SelectedRpsGameDialog
        dialogProps={{ open, onOpenChange }}
        gameId={gameId}
        onClose={() => {
          onClickGameId(null);
        }}
      />
    </div>
  );
}
export enum GameState {
  Win = "Win",
  Lose = "Lose",
  Tie = "Tie",
  Pending = "Pending",
  Submitted = "Submitted",
  Cancelled = "Cancelled",
  Expired = "Expired",
}

export const GameStateClassValMap: Record<GameState, ClassValue> = {
  [GameState.Win]: "bg-green-500",
  [GameState.Lose]: "bg-red-500",
  [GameState.Tie]: "bg-gray-500",
  [GameState.Expired]: "bg-gray-500",
  [GameState.Submitted]: "bg-gray-500",
  [GameState.Pending]: "bg-blue-500",
  [GameState.Cancelled]: "bg-gray-500",
};

function CalculateGameState(game: PlayerRpsGame): GameState {
  const expired = new Date(game.rpsGame.expires_at).getTime() < Date.now();

  if (game.rpsGame.status === "completed") {
    switch (game.player.result) {
      case "tie":
        return GameState.Tie;
      case "win":
        return GameState.Win;
      case "lose":
        return GameState.Lose;
    }
  }
  if (expired) {
    return GameState.Expired;
  }
  if (game.rpsGame.status === "pending") {
    if (game.player.status === "completed") {
      return GameState.Submitted;
    }
    if (game.player.status === "pending") {
      return GameState.Pending;
    }
  } else {
    return GameState.Cancelled;
  }
  return GameState.Pending;
}
