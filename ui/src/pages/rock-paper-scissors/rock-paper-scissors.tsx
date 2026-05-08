import { useOnboardingProgress } from "@/hooks/use-onboarding-progress";
import { rpsGameQueries } from "@/lib/rps-game-queries";
import { RpsGameWithParticipants } from "@/schema.types";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";
import { GameResult } from "../account/rock-paper-scissors/game-result";
import { ErrorCard } from "@/components/error-card";
import { MoveSelection } from "../account/rock-paper-scissors/move";
import { Spinner } from "@/components/ui/spinner";
import { RockPaperScissorsLanding } from "./landing";

export default function RockPaperScissorsPage() {
  const [played, setPlayed] = useState(false);
  const [game, setGame] = useState<RpsGameWithParticipants | null>(null);
  const [searchParams] = useSearchParams();
  const { markStep, progress } = useOnboardingProgress();
  const navigate = useNavigate();
  useEffect(() => {
    const isFirstVisit = !progress.visitedRps;
    markStep("visitedRps");
    if (isFirstVisit) {
      const t = setTimeout(() => {
        toast("Unlock more with a plan!", {
          description: "Subscribe to access protected routes and features.",
          action: { label: "See plans →", onClick: () => navigate("/pricing") },
        });
      }, 2000);
      return () => clearTimeout(t);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  const token = searchParams.get("token");
  const {
    data: rpsGame,
    isLoading,
    error,
  } = useQuery({
    enabled: !!token,
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
  const isSelection = !played && rpsGame?.data;
  const isResult = played && game;
  const isLanding = !isSelection && !isResult;
  if (isLoading) return <Spinner />;
  if (error) return <ErrorCard message={error.message} />;
  return (
    <div className="flex flex-col gap-4 items-center justify-center">
      <div>
        {isLanding && <RockPaperScissorsLanding />}
        {isSelection && (
          <div>
            {rpsGame.data.rps_game.bet_amount ? (
              <div className="mb-4 rounded-lg border border-amber-300 bg-amber-50 p-4 text-center">
                <p className="font-semibold text-amber-800">
                  🪙 This game has a {rpsGame.data.rps_game.bet_amount} pts bet
                </p>
                <p className="mt-1 text-sm text-amber-700">
                  Accepting will deduct {rpsGame.data.rps_game.bet_amount} pts from
                  your balance
                </p>
              </div>
            ) : null}
            <MoveSelection
              handleSubmit={(move) => mutation.mutate({ token: token!, move })}
              opponentPlayer={rpsGame?.data.requesting_participant?.player}
            />
          </div>
        )}
        {isResult && (
          <div>
            <GameResult
              result={game.invited_participant.result}
              opponent={game.requesting_participant.player?.email || ""}
              playerMove={game.invited_participant.move}
              opponentMove={game.requesting_participant.move}
              betAmount={game.rps_game.bet_amount}
              betResult={game.invited_participant.result}
            />
          </div>
        )}
        {played && !rpsGame && <div>Something went wrong...</div>}
      </div>
    </div>
  );
}
