import { AlertCircle } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";

export type ErrorCardProps = {
  exclamation?: string;
  title?: string;
  message?: string;
};

export const ErrorCard = ({
  exclamation = "Oops!",
  title = "Something went wrong",
  message = "We encountered an unexpected error. This could be due to a connection issue or a problem on our end. Please try again",
}: ErrorCardProps) => {
  return (
    <div className="flex flex-col gap-6 items-center justify-center h-screen">
      <div className="max-w-md w-full text-center space-y-6">
        <div className="flex justify-center">
          <div className="rounded-full bg-destructive/10 p-6">
            <AlertCircle className="w-16 h-16 text-destructive" />
          </div>
        </div>

        <div className="space-y-2">
          <h1 className="text-4xl font-bold text-foreground">{exclamation}</h1>
          <p className="text-xl text-muted-foreground">{title}</p>
        </div>

        <p className="text-muted-foreground leading-relaxed">{message}</p>

        <div className="flex flex-col sm:flex-row gap-3 justify-center pt-4">
          <Button asChild size="lg" className="min-w-32">
            <Link to="/">Go Home</Link>
          </Button>
        </div>
      </div>
    </div>
  );
};
