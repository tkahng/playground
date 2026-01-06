import { rpsGameQueries } from "@/lib/rps-game-queries";
import { RpsGameWithParticipants } from "@/schema.types";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useSearchParams } from "react-router";
import { toast } from "sonner";
import { GameResult } from "../account/rock-paper-scissors/game-result";
import { ErrorCard } from "@/components/error-card";
import { MoveSelection } from "../account/rock-paper-scissors/move";

export default function RockPaperScissorsPage() {
  const [played, setPlayed] = useState(false);
  const [game, setGame] = useState<RpsGameWithParticipants | null>(null);
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const {
    data: rpsGame,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["rps-game-with-token", token],
    retry: false,
    queryFn: async () => {
      if (!token) {
        throw new Error("Missing token");
      }
      return rpsGameQueries.getRpsGameWithToken({ token });
    },
  });
  const mutation = useMutation({
    mutationFn: async ({
      token,
      move,
    }: {
      token: string;
      move: "rock" | "paper" | "scissors";
    }) => {
      return rpsGameQueries.submitMoveWithToken({
        token,
        move,
        status: "completed",
      });
    },
    onError(error) {
      toast.error(`Failed to create task: ${error.message}`);
    },
    onSuccess: (data) => {
      setPlayed(true);
      setGame(data.data ?? null);
    },
  });

  if (!token) return <p>Missing token</p>;
  if (isLoading) return <p>Loading...</p>;
  if (error) return <ErrorCard />;
  if (!rpsGame) return <p>Game not found</p>;
  return (
    <div className="flex flex-col gap-4 items-center justify-center">
      <div>
        {!played && (
          <div>
            <MoveSelection
              handleSubmit={(move) => mutation.mutate({ token: token!, move })}
              opponentPlayer={rpsGame?.data.requesting_participant?.player}
            />
          </div>
        )}
        {played && game && (
          <div>
            <GameResult
              {...{
                result: game.invited_participant.result,
                opponent: game.requesting_participant.player?.email || "",
                playerMove: game.invited_participant.move,
                opponentMove: game.requesting_participant.move,
              }}
            />
          </div>
        )}
        {played && !rpsGame && <div>Something went wrong...</div>}
      </div>
    </div>
  );
}
