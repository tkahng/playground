import { useState } from "react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Gamepad2 } from "lucide-react";
import { Player } from "@/schema.types";

export type Move = "rock" | "paper" | "scissors";

const moves = [
  {
    id: "rock" as const,
    emoji: "✊",
    label: "Rock",
    description: "Crushes scissors",
  },
  {
    id: "paper" as const,
    emoji: "✋",
    label: "Paper",
    description: "Covers rock",
  },
  {
    id: "scissors" as const,
    emoji: "✌️",
    label: "Scissors",
    description: "Cuts paper",
  },
];

export type MoveSelectionProps = {
  handleSubmit: (move: Move) => void;
  opponentPlayer?: Player | null;
};

export function MoveSelection({
  handleSubmit,
  opponentPlayer,
}: MoveSelectionProps) {
  const [selectedMove, setSelectedMove] = useState<Move | null>(null);

  return (
    <div className="w-full max-w-4xl animate-in fade-in zoom-in duration-500">
      {/* Header */}
      <div className="text-center mb-8">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary/10 mb-4">
          <Gamepad2 className="w-8 h-8 text-primary" />
        </div>
        <h1 className="text-4xl md:text-5xl font-bold mb-3">
          Choose Your Move
        </h1>
        <p className="text-lg text-muted-foreground">
          Select rock, paper, or scissors to play
        </p>
        {/* Helper Text */}
        {/* <p className="text-center text-sm text-muted-foreground mt-6"> */}
        {/*   Waiting for opponent to join... */}
        {/* </p> */}
        {/* Opponent Info */}
        {opponentPlayer && (
          <div className="text-center mb-6">
            <p className="text-lg text-muted-foreground">VS</p>
            <p className="text-2xl font-semibold">{opponentPlayer?.email}</p>
          </div>
        )}
      </div>

      {/* Move Selection Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 mb-4">
        {moves.map((move) => (
          <button
            key={move.id}
            onClick={() => setSelectedMove(move.id)}
            className={`group relative transition-all duration-200 ${
              selectedMove === move.id ? "scale-105" : "hover:scale-102"
            }`}
          >
            <Card
              className={`p-2 cursor-pointer transition-all duration-200 ${
                selectedMove === move.id
                  ? "border-primary border-2 shadow-lg bg-primary/5"
                  : "border-border hover:border-primary/50 hover:shadow-md"
              }`}
            >
              <div className="flex flex-row sm:flex-col text-center space-y-2 items-center justify-center">
                {/* Emoji */}
                <div
                  className={`text-7xl transition-transform duration-200 ${
                    selectedMove === move.id
                      ? "animate-in zoom-in duration-300"
                      : "group-hover:scale-110"
                  }`}
                >
                  {move.emoji}
                </div>

                {/* Label */}
                <div>
                  <h3 className="text-2xl font-bold mb-1">{move.label}</h3>
                  <p className="text-sm text-muted-foreground">
                    {move.description}
                  </p>
                </div>

                {/* Selected Indicator */}
                {selectedMove === move.id && (
                  <div className="absolute top-4 right-4 w-6 h-6 rounded-full bg-primary flex items-center justify-center animate-in zoom-in duration-200">
                    <svg
                      className="w-4 h-4 text-primary-foreground"
                      fill="none"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                )}
              </div>
            </Card>
          </button>
        ))}
      </div>

      {/* Submit Button */}
      <div className="flex justify-center">
        <Button
          size="lg"
          className="min-w-64 text-lg h-12"
          disabled={!selectedMove}
          onClick={() => handleSubmit(selectedMove || "rock")}
        >
          {selectedMove
            ? `Play ${selectedMove.charAt(0).toUpperCase() + selectedMove.slice(1)}`
            : "Select a Move"}
        </Button>
      </div>
    </div>
  );
}
