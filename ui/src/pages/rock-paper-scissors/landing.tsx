import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { Link } from "@tanstack/react-router";

export function RockPaperScissorsLanding() {
  // Mock authentication state - replace with actual auth
  const authInfo = useAuthProvider();
  const isAuthenticated = authInfo.user;

  return (
    <main className="min-h-screen bg-background">
      {/* Hero Section */}
      <section className="container mx-auto px-4 py-20 md:py-32">
        <div className="max-w-4xl mx-auto text-center space-y-8 animate-in fade-in zoom-in duration-700">
          <div className="inline-block">
            <div className="bg-accent/10 text-accent text-sm font-semibold px-4 py-2 rounded-full mb-6">
              Challenge Players Worldwide
            </div>
          </div>

          <h1 className="text-5xl md:text-7xl font-bold leading-tight text-balance">
            The ultimate Rock Paper Scissors arena
          </h1>

          <p className="text-xl md:text-2xl text-muted-foreground leading-relaxed max-w-2xl mx-auto">
            Battle opponents in real-time, track your wins, and climb the
            leaderboard in the most exciting RPS game online.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center pt-4">
            {isAuthenticated ? (
              <Button asChild size="lg" className="text-lg h-14 px-8">
                <Link to={RouteMap.ACCOUNT_ROCK_PAPER_SCISSORS}>
                  Go play with your friends!
                </Link>
              </Button>
            ) : (
              <Button asChild size="lg" className="text-lg h-14 px-8">
                <Link to={RouteMap.SIGNIN}>Sign in to Play</Link>
              </Button>
            )}
          </div>
        </div>
      </section>

      {/* How to Play Section */}
      <section id="how-to-play" className="container mx-auto px-4 py-20">
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-4xl md:text-5xl font-bold mb-4">How to Play</h2>
            <p className="text-xl text-muted-foreground">
              Master the classic game with these simple rules
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
            <div className="text-center space-y-4">
              <div className="text-7xl mb-2">✊</div>
              <h3 className="text-2xl font-bold">Rock</h3>
              <p className="text-muted-foreground">
                Crushes scissors but loses to paper
              </p>
            </div>

            <div className="text-center space-y-4">
              <div className="text-7xl mb-2">✋</div>
              <h3 className="text-2xl font-bold">Paper</h3>
              <p className="text-muted-foreground">
                Covers rock but loses to scissors
              </p>
            </div>

            <div className="text-center space-y-4">
              <div className="text-7xl mb-2">✌️</div>
              <h3 className="text-2xl font-bold">Scissors</h3>
              <p className="text-muted-foreground">
                Cuts paper but loses to rock
              </p>
            </div>
          </div>

          <Card className="p-8 bg-accent/5 border-accent/20">
            <h3 className="text-2xl font-bold mb-4">Game Rules</h3>
            <ul className="space-y-3 text-lg text-muted-foreground leading-relaxed">
              <li className="flex items-start gap-3">
                <span className="text-accent font-bold mt-1">1.</span>
                <span>Choose your move: rock, paper, or scissors</span>
              </li>
              <li className="flex items-start gap-3">
                <span className="text-accent font-bold mt-1">2.</span>
                <span>Both players reveal their choices simultaneously</span>
              </li>
              <li className="flex items-start gap-3">
                <span className="text-accent font-bold mt-1">3.</span>
                <span>The winner is determined by the classic rules above</span>
              </li>
              <li className="flex items-start gap-3">
                <span className="text-accent font-bold mt-1">4.</span>
                <span>If both players choose the same move, it's a tie</span>
              </li>
            </ul>
          </Card>
        </div>
      </section>

      {/* CTA Section */}
      <section className="container mx-auto px-4 py-20 bg-gradient-to-b from-muted/30 to-background">
        <div className="max-w-3xl mx-auto text-center space-y-8">
          <h2 className="text-4xl md:text-5xl font-bold text-balance">
            Ready to prove your skills?
          </h2>
          <p className="text-xl text-muted-foreground leading-relaxed">
            Join thousands of players competing in the ultimate Rock Paper
            Scissors showdown
          </p>
          <div className="pt-4">
            {isAuthenticated ? (
              <Button asChild size="lg" className="text-lg h-14 px-12">
                <Link to={RouteMap.ACCOUNT_ROCK_PAPER_SCISSORS}>
                  Challenge your friends!
                </Link>
              </Button>
            ) : (
              <Button asChild size="lg" className="text-lg h-14 px-12">
                <Link to={RouteMap.SIGNIN}>Sign in to Play</Link>
              </Button>
            )}
          </div>
        </div>
      </section>
    </main>
  );
}
