import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTrigger } from "@/components/ui/dialog";
import { Field, FieldLabel, FieldError } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { Switch } from "@/components/ui/switch";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { usePlayer } from "@/hooks/use-current-player";
import { useDialog } from "@/hooks/use-dialog";
import { friendsQueries, Friendship } from "@/lib/friends-queries";
import { rpsGameQueries } from "@/lib/rps-game-queries";
import { Player } from "@/schema.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";
import z from "zod";
import { Move, MoveSelection } from "./move";

export const moves = ["rock", "paper", "scissors"] as const;

const searchFormSchema = z.object({
  email: z.string().email(),
});

export type MoveProps = {
  move: Move;
};

function FriendsQuickPick({
  token,
  currentPlayerId,
  onSelect,
}: {
  token?: string;
  currentPlayerId?: string;
  onSelect: (player: Player) => void;
}) {
  const { data, isLoading } = useQuery({
    queryKey: [{ key: "friends" }],
    enabled: !!token,
    queryFn: () => friendsQueries.listFriends({ token: token! }),
  });

  const friends = data?.data ?? [];
  if (isLoading || friends.length === 0) return null;

  const friendPlayers = friends
    .map((f: Friendship) => {
      if (f.requesting_player_id === currentPlayerId) return f.invited_player;
      return f.requesting_player;
    })
    .filter(Boolean) as Player[];

  if (friendPlayers.length === 0) return null;

  return (
    <div>
      <p className="text-xs text-muted-foreground font-medium uppercase tracking-wide mb-1.5">
        Challenge a friend
      </p>
      <div className="flex flex-col gap-1 max-h-32 overflow-y-auto">
        {friendPlayers.map((p) => (
          <button
            key={p.id}
            type="button"
            className="text-left text-sm px-2 py-1.5 rounded hover:bg-accent transition-colors"
            onClick={() => onSelect(p)}
          >
            {p.display_name || p.email}
          </button>
        ))}
      </div>
      <hr className="mt-3" />
    </div>
  );
}

export function CreateGameDialog({
  initialPlayer,
  trigger,
}: {
  initialPlayer?: Player | null;
  trigger?: React.ReactNode;
}) {
  const { user } = useAuthProvider();
  const { player: currentPlayer } = usePlayer();
  const queryClient = useQueryClient();
  const { props: dialogProps } = useDialog();
  const [player, setPlayer] = useState<Player | null>(initialPlayer ?? null);

  const [searched, setSearched] = useState(!!initialPlayer);
  const [emailRequest, setEmailRequest] = useState(false);
  const [email, setEmail] = useState<string | null>(null);
  const [betEnabled, setBetEnabled] = useState(false);
  const [betAmount, setBetAmount] = useState<number | undefined>(undefined);

  const { data: balanceData } = useQuery({
    queryKey: [{ key: "ledger-balance" }],
    enabled: !!user?.tokens.access_token,
    queryFn: async () => {
      if (!user?.tokens.access_token) throw new Error("No access token");
      return rpsGameQueries.getLedgerBalance({ token: user.tokens.access_token });
    },
  });

  const searchForm = useForm<z.infer<typeof searchFormSchema>>({
    resolver: zodResolver(searchFormSchema),
    defaultValues: {
      email: "",
    },
  });

  const emailRequestMutation = useMutation({
    mutationFn: async (data: MoveProps) => {
      if (!user) {
        throw new Error("No user");
      }
      if (!email) {
        throw new Error("no email");
      }
      return rpsGameQueries.requestGameEmail({
        token: user?.tokens.access_token,
        move: data.move,
        invitingPlayerEmail: email,
      });
    },
    onError: (error) => {
      toast.error(error.message);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: [{ key: "rps-games" }],
      });
      toast.success("Game request sent to email");
      dialogProps.onOpenChange(false);
    },
  });

  const requestGameMutation = useMutation({
    mutationFn: async (data: MoveProps) => {
      if (!user) {
        throw new Error("No user");
      }
      if (!player) {
        throw new Error("No player");
      }
      if (betEnabled && (betAmount === undefined || betAmount < 1)) {
        throw new Error("Please enter a valid bet amount");
      }
      return rpsGameQueries.requestGame({
        token: user?.tokens.access_token,
        move: data.move,
        playerId: player.id,
        betAmount: betEnabled ? betAmount : undefined,
      });
    },
    onError: (error) => {
      toast.error(error.message);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: [{ key: "rps-games" }],
      });
      await queryClient.invalidateQueries({
        queryKey: [{ key: "ledger-balance" }],
      });
      toast.success("Game request sent");
      dialogProps.onOpenChange(false);
    },
  });

  const findPlayerMutation = useMutation({
    mutationFn: async (props: z.infer<typeof searchFormSchema>) => {
      if (!user) throw new Error("No user");
      setEmail(props.email);
      // Remove setSearched(true) from here
      return rpsGameQueries.findPlayerByEmail({
        // Return data directly
        token: user.tokens.access_token,
        email: props.email,
      });
    },
    onSuccess: (data) => {
      // Now data is guaranteed
      setSearched(true);
      setPlayer(data?.data ?? null);
    },
    onError: (error) => {
      setSearched(true); // Show error UI too
      toast.error(error.message);
    },
  });

  const onSearchSubmit = (data: z.infer<typeof searchFormSchema>) => {
    findPlayerMutation.mutate(data);
  };

  return (
    <Dialog {...dialogProps}>
      <DialogTrigger asChild>
        {trigger ?? <Button>Play a game with a friend</Button>}
      </DialogTrigger>

      <DialogContent
        onCloseAutoFocus={() => {
          if (!initialPlayer) setPlayer(null);
          setEmailRequest(false);
          if (!initialPlayer) setSearched(false);
          setBetEnabled(false);
          setBetAmount(undefined);
          searchForm.reset();
        }}
      >
        <div className="space-y-4">
          {/* friends quick-pick */}
          {!searched && !emailRequest && !initialPlayer && (
            <FriendsQuickPick
              token={user?.tokens.access_token}
              currentPlayerId={currentPlayer?.id}
              onSelect={(p) => { setPlayer(p as unknown as Player); setSearched(true); }}
            />
          )}

          {/* search input */}

          {!searched && !emailRequest && (
            <>
              {" "}
              <form
                role="search"
                id="search-player-by-email-form"
                onSubmit={searchForm.handleSubmit(onSearchSubmit)}
              >
                <Controller
                  name="email"
                  control={searchForm.control}
                  render={({ field, fieldState }) => (
                    <Field data-invalid={fieldState.invalid}>
                      <FieldLabel htmlFor={field.name}>Email</FieldLabel>
                      <Input
                        {...field}
                        id={field.name}
                        aria-invalid={fieldState.invalid}
                        placeholder="Enter full email address"
                        autoComplete="off"
                        type="search"
                      />
                      {fieldState.invalid && (
                        <FieldError errors={[fieldState.error]} />
                      )}
                    </Field>
                  )}
                />
              </form>
              <Button type="submit" form="search-player-by-email-form">
                Search
              </Button>
            </>
          )}

          {/* searching... */}
          {findPlayerMutation.isPending && <Spinner />}

          {/* error */}
          {searched &&
            !findPlayerMutation.isPending &&
            findPlayerMutation.isError && (
              <div className="rounded-lg border p-4">
                <p>Error: {findPlayerMutation.error.message}</p>
              </div>
            )}

          {/* not found */}
          {searched &&
            !player &&
            !findPlayerMutation.isPending &&
            !findPlayerMutation.isError && (
              <>
                <div className="rounded-lg border p-4">
                  <p>No user found.</p>
                  <p> Would you like to send game request instead?</p>
                </div>
                <div className="flex flex-row gap-2 items-center justify-around">
                  <Button variant="ghost" onClick={() => setSearched(false)}>
                    Search Again
                  </Button>
                  <Button
                    onClick={() => {
                      setSearched(false);
                      setEmailRequest(true);
                    }}
                  >
                    Send game request
                  </Button>
                </div>
              </>
            )}

          {/* send request */}
          {!searched &&
            !player &&
            emailRequest &&
            email &&
            !findPlayerMutation.isPending &&
            !findPlayerMutation.isError && (
              <MoveSelection
                handleSubmit={(move) =>
                  emailRequestMutation.mutate({
                    move,
                  })
                }
              />
            )}

          {/* user found. select move */}
          {searched &&
            player &&
            !findPlayerMutation.isPending &&
            !findPlayerMutation.isError && (
              <MoveSelection
                handleSubmit={(move: Move) =>
                  requestGameMutation.mutate({ move })
                }
              >
                <div className="border-t pt-3 mt-2 space-y-2">
                  <div className="flex items-center justify-between">
                    <label htmlFor="bet-toggle" className="text-sm cursor-pointer">
                      Add a bet?
                    </label>
                    <Switch
                      id="bet-toggle"
                      checked={betEnabled}
                      onCheckedChange={(checked) => {
                        setBetEnabled(checked);
                        if (!checked) setBetAmount(undefined);
                      }}
                    />
                  </div>
                  {betEnabled && (
                    <div className="space-y-1">
                      {(balanceData === undefined || (balanceData.available_balance ?? 0) > 0) && (
                        <p className="text-xs text-muted-foreground">
                          Available balance:{" "}
                          {balanceData?.available_balance !== undefined
                            ? `${balanceData.available_balance} pts`
                            : "..."}
                        </p>
                      )}
                      {balanceData === undefined ? null : (balanceData.available_balance ?? 0) <= 0 ? (
                        <p className="text-xs text-amber-600">
                          You have 0 pts.{" "}
                          <a
                            href={RouteMap.POINTS_SETTINGS}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="underline hover:text-amber-700"
                          >
                            Buy Points →
                          </a>
                        </p>
                      ) : (
                        <Input
                          type="number"
                          min={1}
                          max={balanceData?.available_balance}
                          value={betAmount ?? ""}
                          onChange={(e) => {
                            const parsed = parseInt(e.target.value, 10);
                            setBetAmount(
                              e.target.value && Number.isFinite(parsed) ? parsed : undefined
                            );
                          }}
                          placeholder="Enter bet amount"
                        />
                      )}
                    </div>
                  )}
                </div>
              </MoveSelection>
            )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
