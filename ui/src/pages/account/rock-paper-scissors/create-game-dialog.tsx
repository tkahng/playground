import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { rpsGameQueries } from "@/lib/rps-game-queries";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useState } from "react";
type Move = "rock" | "paper" | "scissors";
export const moves = ["rock", "paper", "scissors"] as const;
type Props = {
  onPick: (move: Move) => void;
};

export function MovePicker({ onPick }: Props) {
  return (
    <div className="flex gap-2">
      <DialogClose asChild>
        <Button onClick={() => onPick("rock")}>Rock</Button>
      </DialogClose>
      <DialogClose>
        <Button onClick={() => onPick("paper")}>Paper</Button>
      </DialogClose>
      <DialogClose>
        <Button onClick={() => onPick("scissors")}>Scissors</Button>
      </DialogClose>
    </div>
  );
}

export function CreateGameDialog() {
  const { user } = useAuthProvider();
  const [email, setEmail] = useState("");
  const [inputEmail, setInputEmail] = useState(email);

  const [searched, setSearched] = useState(false);

  const searchQuery = useQuery({
    queryKey: [{ key: "search-player-by-email", email: inputEmail }],
    queryFn: () => {
      if (!user) {
        throw new Error("No user");
      }
      return rpsGameQueries.findPlayerByEmail({
        token: user.tokens.access_token,
        email: inputEmail,
      });
    },

    enabled: false,
    retry: false,
  });

  const onSearch = async () => {
    setSearched(true);
    setInputEmail(email);
    await searchQuery.refetch();
  };

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button>Play a game with a friend</Button>
      </DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>Search for a friend</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <Input
            placeholder="Enter full email address"
            value={email}
            type="email"
            onChange={(e) => setEmail(e.target.value)}
          />

          <Button
            onClick={onSearch}
            disabled={!email || searchQuery.isFetching}
          >
            Search
          </Button>

          {searchQuery.isFetching && <Spinner />}

          {searched && !searchQuery.isFetching && searchQuery.isError && (
            <div className="rounded-lg border p-4">
              <p>Error: {searchQuery.error.message}</p>
            </div>
          )}

          {searched &&
            !searchQuery.data &&
            !searchQuery.isFetching &&
            !searchQuery.isError && (
              <div className="rounded-lg border p-4">
                <p>
                  No user found for <strong>{email}</strong>
                </p>
              </div>
            )}

          {searchQuery.data?.data && (
            <div className="rounded-lg border p-4 flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">
                  {searchQuery.data.data.email}
                </p>
              </div>
              <SendGameRequestDialog playerId={searchQuery.data.data.id} />
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

export function SendGameRequestDialog(props: { playerId: string }) {
  const { user } = useAuthProvider();
  const addFriendMutation = useMutation({
    mutationFn: async ({
      playerId,
      token,
      move,
    }: {
      playerId: string;
      token: string;
      move: "rock" | "paper" | "scissors";
    }) => {
      return rpsGameQueries.requestGame({
        token,
        move,
        playerId,
      });
    },
  });
  const onPick = (move: Move) => {
    addFriendMutation.mutate({
      move,
      playerId: props.playerId,
      token: user!.tokens.access_token,
    });
  };
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button>Play a game with a friend</Button>
      </DialogTrigger>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>Choose your move</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <MovePicker onPick={onPick} />
        </div>
      </DialogContent>
    </Dialog>
  );
}
