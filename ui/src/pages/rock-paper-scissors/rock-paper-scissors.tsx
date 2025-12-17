import { Button } from "@/components/ui/button";
import { rpsGameQueries } from "@/lib/rps-game-queries";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useSearchParams } from "react-router";
import { toast } from "sonner";

const moves = ["rock", "paper", "scissors"] as const;

export default function RockPaperScissors() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const {
    data: rpsGame,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["rps-game-with-token", token],
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
  });
  if (!token) return <p>Missing token</p>;
  if (isLoading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;
  if (!rpsGame) return <p>Game not found</p>;
  return (
    <div>
      <h1>Rock Paper Scissors</h1>
      <div>
        {moves.map((move) => (
          <Button
            key={move}
            onClick={() => {
              mutation.mutate({ token: token, move });
            }}
          >
            {move}
          </Button>
        ))}
      </div>
    </div>
  );
}
