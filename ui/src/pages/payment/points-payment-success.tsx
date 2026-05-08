import { PlaygroundLandingHeader } from "@/components/playground-landing-header";
import { PlaygroundMinimalFooter } from "@/components/playground-minimal-footer";
import { RouteMap } from "@/components/route-map";
import { CheckCircle } from "lucide-react";
import { useEffect } from "react";
import { useNavigate } from "@tanstack/react-router";

const REDIRECT_DELAY_MS = 3000;

export default function PointsPaymentSuccessPage() {
  const navigate = useNavigate();

  useEffect(() => {
    const timer = setTimeout(() => {
      navigate(RouteMap.POINTS_SETTINGS);
    }, REDIRECT_DELAY_MS);
    return () => clearTimeout(timer);
  }, [navigate]);

  return (
    <div className="flex flex-col min-h-screen">
      <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
        <PlaygroundLandingHeader />
      </div>
      <main className="flex-1 flex items-center justify-center p-4">
        <div className="max-w-md w-full space-y-6 text-center">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-green-100 dark:bg-green-900">
            <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-300" />
          </div>
          <div>
            <h1 className="text-3xl font-bold">Points Purchased!</h1>
            <p className="text-muted-foreground mt-2">
              Thanks for your purchase. Your points will be credited to your
              account shortly.
            </p>
          </div>
          <p className="text-sm text-muted-foreground">
            Redirecting to your points settings...
          </p>
        </div>
      </main>
      <PlaygroundMinimalFooter />
    </div>
  );
}
