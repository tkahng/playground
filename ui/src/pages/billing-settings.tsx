import CustomerPortalForm from "@/components/customer-portal-form";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getUserSubscriptions } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";

export default function BillingSettingPage() {
  const { user } = useAuthProvider();
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
  if (!data) {
    return <div>No data</div>;
  }
  return (
    <div className="space-y-6 flex w-full flex-col items-center justify-center">
      {/* <div>
        <h3 className="text-lg font-medium">Profile</h3>
        <p className="text-sm text-muted-foreground">
          This is how others will see you on the site.
        </p>
      </div> */}
      <CustomerPortalForm subscription={data} />
    </div>
  );
}
