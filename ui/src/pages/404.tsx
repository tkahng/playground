import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ArrowLeft, Brain, HelpCircle, Home, Search } from "lucide-react";
import { Link } from "react-router";

export default function NotFoundPage() {
  return (
    // <div className="flex flex-col min-h-screen">
    //   <NexusAILandingHeader />
    //   <main className="flex-1 flex items-center justify-center p-4">
    <div className="max-w-md w-full text-center space-y-8">
      <div className="space-y-2">
        <h1 className="text-9xl font-bold text-primary">404</h1>
        <h2 className="text-3xl font-bold tracking-tight">Page not found</h2>
        <p className="text-muted-foreground">
          The page you're looking for doesn't exist or has been moved.
        </p>
      </div>

      <div className="mx-auto max-w-xs">
        <div className="relative">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Search for something..."
            className="pl-8"
          />
        </div>
      </div>

      <div className="flex flex-col sm:flex-row gap-2 justify-center">
        <Button variant="outline" asChild>
          <Link to="javascript:history.back()">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Go Back
          </Link>
        </Button>
        <Button variant="outline" asChild>
          <Link to="/">
            <Home className="mr-2 h-4 w-4" />
            Home
          </Link>
        </Button>
        <Button asChild>
          <Link to="/contact">
            <HelpCircle className="mr-2 h-4 w-4" />
            Get Help
          </Link>
        </Button>
      </div>

      <div className="pt-8">
        <div className="relative mx-auto w-64 h-64 md:w-80 md:h-80">
          <div className="absolute inset-0 bg-gradient-to-r from-primary/20 to-primary/40 rounded-full blur-3xl opacity-50"></div>
          <div className="relative flex items-center justify-center w-full h-full">
            <Brain className="h-32 w-32 text-primary opacity-80" />
          </div>
        </div>
      </div>
    </div>
    //   </main>
    //   <NexusAIMinimalFooter />
    // </div>
  );
}
