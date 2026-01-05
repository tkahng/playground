import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Trophy, XCircle, Minus } from "lucide-react";

type Move = "rock" | "paper" | "scissors";
type Result = "win" | "lose" | "tie";

interface GameResultProps {
  result: Result;
  opponent: string;
  playerMove: Move;
  opponentMove: Move;
}

const moveEmojis: Record<Move, string> = {
  rock: "✊",
  paper: "✋",
  scissors: "✌️",
};

const resultConfig = {
  win: {
    title: "Victory!",
    subtitle: "You won this round",
    icon: Trophy,
    colorClass: "text-success",
    bgClass: "bg-success/10",
  },
  lose: {
    title: "Defeat",
    subtitle: "Better luck next time",
    icon: XCircle,
    colorClass: "text-destructive",
    bgClass: "bg-destructive/10",
  },
  tie: {
    title: "It's a Tie!",
    subtitle: "Great minds think alike",
    icon: Minus,
    colorClass: "text-warning",
    bgClass: "bg-warning/10",
  },
};

export function GameResult({
  result,
  opponent,
  playerMove,
  opponentMove,
}: GameResultProps) {
  const config = resultConfig[result];
  const Icon = config.icon;

  return (
    <div className="flex flex-col gap-6 items-center justify-center h-screen">
      <div className="w-full max-w-2xl animate-in fade-in zoom-in duration-500">
        {/* Result Header */}
        <div className="text-center mb-8">
          <div
            className={`inline-flex items-center justify-center w-20 h-20 rounded-full ${config.bgClass} mb-4`}
          >
            <Icon className={`w-10 h-10 ${config.colorClass}`} />
          </div>
          <h1
            className={`text-5xl md:text-6xl font-bold mb-2 ${config.colorClass}`}
          >
            {config.title}
          </h1>
          <p className="text-lg text-muted-foreground">{config.subtitle}</p>
        </div>

        {/* Opponent Info */}
        <div className="text-center mb-6">
          <p className="text-sm text-muted-foreground mb-1">Played against</p>
          <p className="text-2xl font-semibold">{opponent}</p>
        </div>

        {/* Moves Display */}
        <Card className="p-8 mb-8">
          <div className="grid grid-cols-3 gap-4 items-center">
            {/* Player Move */}
            <div className="text-center">
              <p className="text-sm font-medium text-muted-foreground mb-3">
                You
              </p>
              <div className="text-7xl mb-3 animate-in zoom-in duration-300 delay-100">
                {moveEmojis[playerMove]}
              </div>
              <p className="text-lg font-semibold capitalize">{playerMove}</p>
            </div>

            {/* VS Divider */}
            <div className="flex items-center justify-center">
              <div className="text-2xl font-bold text-muted-foreground">VS</div>
            </div>

            {/* Opponent Move */}
            <div className="text-center">
              <p className="text-sm font-medium text-muted-foreground mb-3">
                Opponent
              </p>
              <div className="text-7xl mb-3 animate-in zoom-in duration-300 delay-200">
                {moveEmojis[opponentMove]}
              </div>
              <p className="text-lg font-semibold capitalize">{opponentMove}</p>
            </div>
          </div>
        </Card>

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row gap-3">
          <Button size="lg" className="flex-1 text-lg" variant="default">
            Play Again
          </Button>
          <Button
            size="lg"
            className="flex-1 text-lg bg-transparent"
            variant="outline"
          >
            View Stats
          </Button>
        </div>
      </div>
    </div>
  );
}
