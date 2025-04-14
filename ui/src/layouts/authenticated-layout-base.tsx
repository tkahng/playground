import { landingLinks } from "@/components/landing-links";
import { NexusAIFooter } from "@/components/nexus-footer";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect } from "react";
import { Navigate, useLocation, useOutlet } from "react-router";

export default function AuthenticatedLayoutBase() {
  const location = useLocation();
  const { checkAuth, user } = useAuthProvider();
  const outlet = useOutlet();
  useEffect(() => {
    checkAuth();
  }, [location]);

  if (!user) {
    return <Navigate to="/signin" />;
  }
  return (
    <>
      <div className="relative flex min-h-screen flex-col justify-center">
        <NexusAILandingHeader leftLinks={landingLinks} />
        <main className="flex flex-grow">
          {/* <DashboardSidebar /> */}
          {outlet}
        </main>
        <NexusAIFooter />
      </div>
    </>
  );
}
