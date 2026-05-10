import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useDialog } from "@/hooks/use-dialog";
import { ChallengeHouseResult, rpsGameQueries } from "@/lib/rps-game-queries";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";
import { Link } from "@tanstack/react-router";
import { Move, MoveSelection } from "./move";
import { GameResult } from "./game-result";

const HOUSE_MAX_BET = 500;

export function ChallengeHouseDialog({ trigger }: { trigger?: React.ReactNode }) {
  const { user } = useAuthProvider();
  const queryClient = useQueryClient();
  const { props: dialogProps } = useDialog();

  const [betEnabled, setBetEnabled] = useState(false);
  const [betAmount, setBetAmount] = useState<number | undefined>(undefined);
  const [result, setResult] = useState<ChallengeHouseResult | null>(null);

  const { data: balanceData } = useQuery({
    queryKey: [{ key: "ledger-balance" }],
    enabled: !!user?.tokens.access_token,
    queryFn: async () => {
      if (!user?.tokens.access_token) throw new Error("No access token");
      return rpsGameQueries.getLedgerBalance({ token: user.tokens.access_token });
    },
  });

  const challengeMutation = useMutation({
    mutationFn: async (move: Move) => {
      if (!user) throw new Error("No user");
      if (betEnabled && (betAmount === undefined || betAmount < 1)) {
        throw new Error("Enter a valid bet amount");
      }
      if (betEnabled && betAmount !== undefined && betAmount > HOUSE_MAX_BET) {
        throw new Error(`Bet cannot exceed ${HOUSE_MAX_BET} pts vs the house`);
      }
      return rpsGameQueries.challengeHouse({
        token: user.tokens.access_token,
        move,
        betAmount: betEnabled ? betAmount : undefined,
      });
    },
    onError: (err) => {
      toast.error(err.message);
    },
    onSuccess: async (data) => {
      setResult(data);
      await queryClient.invalidateQueries({ queryKey: [{ key: "rps-games" }] });
      await queryClient.invalidateQueries({ queryKey: [{ key: "ledger-balance" }] });
    },
  });

  const reset = () => {
    setResult(null);
    setBetEnabled(false);
    setBetAmount(undefined);
    challengeMutation.reset();
  };

  const userMove = result?.requesting_participant.move as Move | undefined;
  const houseMove = result?.invited_participant.move as Move | undefined;
  const userResult = result?.requesting_participant.result as "win" | "lose" | "tie" | undefined;

  return (
    <Dialog {...dialogProps}>
      <DialogTrigger asChild>
        {trigger ?? (
          <Button variant="outline">
            🏠 Challenge the House
          </Button>
        )}
      </DialogTrigger>

      <DialogContent
        onCloseAutoFocus={reset}
        className="max-w-xl"
      >
        {/* Loading state */}
        {challengeMutation.isPending && (
          <div className="flex flex-col items-center justify-center py-12 gap-4 animate-in fade-in duration-300">
            <div className="text-5xl animate-bounce">🏠</div>
            <p className="text-lg font-semibold text-muted-foreground">
              The house is thinking…
            </p>
          </div>
        )}

        {/* Result state */}
        {result && userMove && houseMove && userResult && (
          <div className="space-y-4">
            <GameResult
              result={userResult}
              opponent="🏠 The House"
              playerMove={userMove}
              opponentMove={houseMove}
              betAmount={betEnabled ? betAmount : undefined}
              betResult={userResult}
            />

            {/* "House always wins." catchphrase */}
            {result.house_message && (
              <p className="text-center text-sm font-semibold text-destructive tracking-wide uppercase">
                {result.house_message}
              </p>
            )}

            {/* Cooldown notice */}
            <p className="text-center text-xs text-muted-foreground">
              Next challenge available after{" "}
              {new Date(result.cooldown_ends_at).toLocaleTimeString()}
            </p>

            <div className="flex justify-center">
              <Button variant="outline" onClick={reset}>
                Play again later
              </Button>
            </div>
          </div>
        )}

        {/* Move selection state */}
        {!challengeMutation.isPending && !result && (
          <MoveSelection
            handleSubmit={(move) => challengeMutation.mutate(move)}
            disabled={challengeMutation.isPending}
          >
            <div className="border-t pt-3 mt-2 space-y-2">
              <p className="text-xs text-muted-foreground text-center">
                🏠 The house plays randomly. Results are instant.
              </p>

              <div className="flex items-center justify-between">
                <label htmlFor="house-bet-toggle" className="text-sm cursor-pointer">
                  Add a bet? (max {HOUSE_MAX_BET} pts)
                </label>
                <Switch
                  id="house-bet-toggle"
                  checked={betEnabled}
                  onCheckedChange={(checked) => {
                    setBetEnabled(checked);
                    if (!checked) setBetAmount(undefined);
                  }}
                />
              </div>

              {betEnabled && (
                <div className="space-y-1">
                  {balanceData === undefined ? null : (balanceData.available_balance ?? 0) <= 0 ? (
                    <p className="text-xs text-amber-600">
                      You have 0 pts.{" "}
                      <Link
                        to={RouteMap.POINTS_SETTINGS}
                        className="underline hover:text-amber-700"
                      >
                        Buy Points →
                      </Link>
                    </p>
                  ) : (
                    <>
                      <p className="text-xs text-muted-foreground">
                        Available:{" "}
                        {balanceData.available_balance !== undefined
                          ? `${balanceData.available_balance} pts`
                          : "…"}
                      </p>
                      <Input
                        type="number"
                        min={1}
                        max={Math.min(
                          HOUSE_MAX_BET,
                          balanceData?.available_balance ?? HOUSE_MAX_BET,
                        )}
                        value={betAmount ?? ""}
                        onChange={(e) => {
                          const parsed = parseInt(e.target.value, 10);
                          setBetAmount(
                            e.target.value && Number.isFinite(parsed)
                              ? Math.min(parsed, HOUSE_MAX_BET)
                              : undefined,
                          );
                        }}
                        placeholder={`1 – ${HOUSE_MAX_BET} pts`}
                      />
                    </>
                  )}
                </div>
              )}
            </div>
          </MoveSelection>
        )}
      </DialogContent>
    </Dialog>
  );
}
