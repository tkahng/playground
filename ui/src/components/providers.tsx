import { AuthProvider } from "@/context/auth-context";
import { TeamProvider } from "@/context/team-context";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "./theme-provider";
import { Toaster } from "./ui/sonner";
const queryClient = new QueryClient();

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
          <AuthProvider>
            <TeamProvider>{children}</TeamProvider>
            {/* {children} */}
          </AuthProvider>
          <Toaster />
        </ThemeProvider>
      </QueryClientProvider>
    </>
  );
}
