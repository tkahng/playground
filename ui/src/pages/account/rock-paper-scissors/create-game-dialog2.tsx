import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Field,
  FieldLabel,
  FieldError,
  FieldGroup,
  FieldContent,
  FieldDescription,
  FieldSet,
  FieldTitle,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { RadioGroupItem, RadioGroup } from "@/components/ui/radio-group";
import { Spinner } from "@/components/ui/spinner";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useDialog } from "@/hooks/use-dialog";
import { rpsGameQueries } from "@/lib/rps-game-queries";
import { Player } from "@/schema.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";
import z from "zod";

export const moves = ["rock", "paper", "scissors"] as const;

const searchFormSchema = z.object({
  email: z.string().email(),
});

const requestGameFormSchema = z.object({
  move: z.enum(moves),
});

export function CreateGameDialog() {
  const { user } = useAuthProvider();
  const queryClient = useQueryClient();
  const { props: dialogProps } = useDialog();
  const [player, setPlayer] = useState<Player | null>(null);

  const [searched, setSearched] = useState(false);
  const [emailRequest, setEmailRequest] = useState(false);
  const [email, setEmail] = useState<string | null>(null);
  const searchForm = useForm<z.infer<typeof searchFormSchema>>({
    resolver: zodResolver(searchFormSchema),
    defaultValues: {
      email: "",
    },
  });

  const requestGameForm = useForm<z.infer<typeof requestGameFormSchema>>({
    resolver: zodResolver(requestGameFormSchema),
    defaultValues: {
      move: "rock",
    },
  });

  const emailRequestForm = useForm<z.infer<typeof requestGameFormSchema>>({
    resolver: zodResolver(requestGameFormSchema),
    defaultValues: {
      move: "rock",
    },
  });

  const emailRequestMutation = useMutation({
    mutationFn: async (data: z.infer<typeof requestGameFormSchema>) => {
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

  const onEmailRequestSubmit = (
    data: z.infer<typeof requestGameFormSchema>,
  ) => {
    emailRequestMutation.mutate(data);
  };
  const requestGameMutation = useMutation({
    mutationFn: async (data: z.infer<typeof requestGameFormSchema>) => {
      if (!user) {
        throw new Error("No user");
      }
      if (!player) {
        throw new Error("No player");
      }
      return rpsGameQueries.requestGame({
        token: user?.tokens.access_token,
        move: data.move,
        playerId: player.id,
      });
    },
    onError: (error) => {
      toast.error(error.message);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: [{ key: "rps-games" }],
      });
      toast.success("Game request sent");
      dialogProps.onOpenChange(false);
    },
  });

  const onRequestSubmit = (data: z.infer<typeof requestGameFormSchema>) => {
    requestGameMutation.mutate(data);
  };

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
        <Button>Play a game with a friend</Button>
      </DialogTrigger>

      <DialogContent
        onCloseAutoFocus={() => {
          setPlayer(null);
          setEmailRequest(false);
          setSearched(false);
          searchForm.reset();
        }}
      >
        <DialogHeader>
          <DialogTitle>Search for a friend</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
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
              <>
                <div className="flex flex-col justify-center">
                  <form
                    id="request-game-email-form"
                    className="rounded-lg border p-4 space-y-2 flex flex-col items-center justify-center"
                    onSubmit={emailRequestForm.handleSubmit(
                      onEmailRequestSubmit,
                    )}
                  >
                    <p className="text-lg font-bold">{email}</p>

                    <FieldGroup className="flex flex-col gap-2 items-center justify-center">
                      <Controller
                        name="move"
                        control={emailRequestForm.control}
                        render={({ field, fieldState }) => (
                          <FieldSet data-invalid={fieldState.invalid}>
                            <FieldDescription>
                              Choose your move
                            </FieldDescription>
                            <RadioGroup
                              name={field.name}
                              value={field.value}
                              className="flex flex-col sm:flex-row"
                              onValueChange={field.onChange}
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
                  </form>
                  <Button type="submit" form="request-game-email-form">
                    Send game request
                  </Button>
                </div>
              </>
            )}

          {/* user found. select move */}
          {searched &&
            player &&
            !findPlayerMutation.isPending &&
            !findPlayerMutation.isError && (
              <div>
                <form
                  id="request-game-form"
                  className="rounded-lg border p-4 space-y-2 flex flex-col items-center justify-center"
                  onSubmit={requestGameForm.handleSubmit(onRequestSubmit)}
                >
                  <p className="text-lg font-bold">{player.email}</p>

                  <FieldGroup className="flex flex-col gap-2 items-center justify-center">
                    <Controller
                      name="move"
                      control={requestGameForm.control}
                      render={({ field, fieldState }) => (
                        <FieldSet data-invalid={fieldState.invalid}>
                          <FieldDescription>Choose your move</FieldDescription>
                          <RadioGroup
                            name={field.name}
                            value={field.value}
                            className="flex flex-col sm:flex-row"
                            onValueChange={field.onChange}
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
                </form>
                <Button type="submit" form="request-game-form">
                  Send game request
                </Button>
              </div>
            )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
