import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { settingsSidebarLinks } from "@/components/links";
import TeamCustomerForm from "@/components/team-customer-form";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getUserSubscriptions } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";

export default function BillingSettingPage() {
  const { user } = useAuthProvider();
  const { data, isPending } = useQuery({
    queryKey: ["billing-settings"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      try {
        return getUserSubscriptions(user.tokens.access_token);
      } catch {
        return null;
      }
    },
    retry: false,
  });
  if (isPending) {
    return <div>Loading...</div>;
  }

  return (
    <div className="flex">
      <DashboardSidebar links={settingsSidebarLinks} />
      <div className="flex-1 space-y-6 p-12 w-full">
        <TeamCustomerForm subscription={data || null} />
      </div>
    </div>
  );
}
