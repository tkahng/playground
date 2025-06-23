import CustomerPortalForm from "@/components/customer-portal-form";
import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { teamSettingLinks } from "@/components/links";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeamContext } from "@/hooks/use-team-context";
import { getUserSubscriptions } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";

export default function TeamBillingSettingPage() {
  const { user } = useAuthProvider();
  const { team } = useTeamContext();
  const { data, isPending, isError, error } = useQuery({
    queryKey: ["billing-settings"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      return getUserSubscriptions(user.tokens.access_token);
    },
  });
  if (isPending) {
    return <div>Loading...</div>;
  }

  if (isError) {
    return <div>Error: {error.message}</div>;
  }
  if (!team?.id) {
    return <div>Error: Team ID is required</div>;
  }
  return (
    <div className="flex">
      <DashboardSidebar links={teamSettingLinks(team.slug)} />
      <div className="flex-1 space-y-6 p-12 w-full">
        <CustomerPortalForm subscription={data} />
      </div>
    </div>
  );
}
