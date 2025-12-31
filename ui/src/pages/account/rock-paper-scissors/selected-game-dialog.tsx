import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import {
  FieldGroup,
  FieldSet,
  FieldDescription,
  FieldLabel,
  Field,
  FieldContent,
  FieldTitle,
  FieldError,
} from "@/components/ui/field";
import { RadioGroupItem, RadioGroup } from "@/components/ui/radio-group";
import { rpsGameQueries } from "@/lib/rps-game-queries";
import { PlayerRpsGame } from "@/schema.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm, Controller } from "react-hook-form";
import { toast } from "sonner";
import z from "zod";
import { moves } from "./create-game-dialog";
import { useAuthProvider } from "@/hooks/use-auth-provider";

export const SelectedRpsGameDialog = ({
  dialogProps,
  onClose,
  rpsGame,
}: {
  dialogProps: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
  };
  onClose: () => void;
  rpsGame: PlayerRpsGame | null;
}) => {
  const gameTime = new Date(rpsGame?.rpsGame.expires_at || 0);
  // eslint-disable-next-line react-hooks/purity
  const expired = gameTime.getTime() < Date.now();
  return (
    <Dialog {...dialogProps}>
      <DialogContent
        onCloseAutoFocus={() => {
          onClose();
        }}
      >
        <DialogTitle>
          Your game against{" "}
          {rpsGame?.opponent.player?.display_name ||
            rpsGame?.opponent.player?.email}
        </DialogTitle>

        <div className="flex flex-col gap-4">
          {/* no game found */}
          {!rpsGame && <NoPlayerView onOpenChange={dialogProps.onOpenChange} />}

          {/* game found with pending status and you have submitted your move */}
          {rpsGame &&
            !expired &&
            rpsGame.rpsGame.status === "pending" &&
            rpsGame.player.status === "completed" && (
              <PendingGameView
                onOpenChange={dialogProps.onOpenChange}
                game={rpsGame}
              />
            )}

          {/* game found with pending status and you have not submitted your move */}
          {rpsGame &&
            !expired &&
            rpsGame.rpsGame.status === "pending" &&
            rpsGame.player.status === "pending" && (
              <SubmitMoveView
                onOpenChange={dialogProps.onOpenChange}
                game={rpsGame}
              />
            )}
          {rpsGame && expired && (
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

const submitToGameFormSchema = z.object({
  move: z.enum(moves),
});

export const SubmitMoveView = ({
  onOpenChange,
  game,
}: {
  onOpenChange: (open: boolean) => void;
  game: PlayerRpsGame;
}) => {
  const { user } = useAuthProvider();
  const queryClient = useQueryClient();
  const submitToGameForm = useForm<z.infer<typeof submitToGameFormSchema>>({
    resolver: zodResolver(submitToGameFormSchema),
    defaultValues: {
      move: "rock",
    },
  });

  const submitToGameMutation = useMutation({
    mutationFn: async (data: z.infer<typeof submitToGameFormSchema>) => {
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
        queryKey: [{ key: "rps-games" }],
      });
      toast.success("move submitted");
      onOpenChange(false);
    },
  });

  const onRequestSubmit = (data: z.infer<typeof submitToGameFormSchema>) => {
    submitToGameMutation.mutate(data);
  };
  return (
    <div>
      <form
        id="submit-game-form"
        className="rounded-lg border p-4 space-y-2 flex flex-col items-center justify-center"
        onSubmit={submitToGameForm.handleSubmit(onRequestSubmit)}
      >
        <p className="text-lg font-bold"> Submit your move</p>

        <FieldGroup className="flex flex-col gap-2 items-center justify-center">
          <Controller
            name="move"
            control={submitToGameForm.control}
            render={({ field, fieldState }) => (
              <FieldSet data-invalid={fieldState.invalid}>
                <FieldDescription>Choose your move</FieldDescription>
                <RadioGroup
                  name={field.name}
                  value={field.value}
                  onValueChange={field.onChange}
                  className="flex flex-col sm:flex-row"
                  aria-invalid={fieldState.invalid}
                >
                  {moves.map((m) => (
                    <FieldLabel key={m} htmlFor={`form-rhf-radiogroup-${m}`}>
                      <Field
                        orientation="horizontal"
                        data-invalid={fieldState.invalid}
                      >
                        <FieldContent>
                          <FieldTitle className="text-lg font-bold">
                            {m}
                          </FieldTitle>
                        </FieldContent>
                        <RadioGroupItem
                          value={m}
                          id={`form-rhf-radiogroup-${m}`}
                          aria-invalid={fieldState.invalid}
                        />
                      </Field>
                    </FieldLabel>
                  ))}
                </RadioGroup>
                {fieldState.invalid && (
                  <FieldError errors={[fieldState.error]} />
                )}
              </FieldSet>
            )}
          />
        </FieldGroup>
        <Button type="submit" form="submit-game-form">
          Submit move
        </Button>
      </form>
    </div>
  );
};
