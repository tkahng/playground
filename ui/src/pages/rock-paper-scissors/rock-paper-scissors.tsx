import { Button } from "@/components/ui/button";
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
import { RpsGameWithParticipants } from "@/schema.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useForm, Controller } from "react-hook-form";
import { useSearchParams } from "react-router";
import { toast } from "sonner";
import z from "zod";
import { GameResult } from "../account/rock-paper-scissors/game-result";
import { ErrorCard } from "@/components/error-card";

const moves = ["rock", "paper", "scissors"] as const;

const submitToGameFormSchema = z.object({
  move: z.enum(moves),
});

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

  const submitToGameForm = useForm<z.infer<typeof submitToGameFormSchema>>({
    resolver: zodResolver(submitToGameFormSchema),
    defaultValues: { move: "rock" },
  });
  const onSubmit = (data: z.infer<typeof submitToGameFormSchema>) => {
    mutation.mutate({ token: token!, move: data.move });
  };
  if (!token) return <p>Missing token</p>;
  if (isLoading) return <p>Loading...</p>;
  if (error) return <ErrorCard />;
  if (!rpsGame) return <p>Game not found</p>;
  return (
    <div>
      <h1>Rock Paper Scissors</h1>
      <div>
        {!played && (
          <div>
            <form
              id="submit-game-form"
              className="rounded-lg border p-4 space-y-2 flex flex-col items-center justify-center"
              onSubmit={submitToGameForm.handleSubmit(onSubmit)}
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
                          <FieldLabel
                            key={m}
                            htmlFor={`form-rhf-radiogroup-${m}`}
                          >
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
        )}
        {played && game && (
          <GameResult
            {...{
              result: game.invited_participant.result,
              opponent: game.requesting_participant.player?.email || "",
              playerMove: game.invited_participant.move,
              opponentMove: game.requesting_participant.move,
            }}
          ></GameResult>
        )}
        {played && !rpsGame && <div>Something went wrong...</div>}
      </div>
    </div>
  );
}
