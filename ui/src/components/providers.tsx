import { AuthProvider } from "@/context/auth-context";
import { TeamProvider2 } from "@/context/team-context2";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { ThemeProvider } from "./theme-provider";
import { Toaster } from "./ui/sonner";

const queryClient = new QueryClient();

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
          <AuthProvider>
            {/* <TeamProvider> */}
            <TeamProvider2>{children}</TeamProvider2>
            {/* </TeamProvider> */}
            {/* {children} */}
          </AuthProvider>
          <Toaster />
        </ThemeProvider>
        {import.meta.env.DEV && <ReactQueryDevtools />}
      </QueryClientProvider>
    </>
  );
}
