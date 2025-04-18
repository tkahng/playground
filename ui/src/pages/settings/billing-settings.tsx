import CustomerPortalForm from "@/components/customer-portal-form";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getUserSubscriptions } from "@/lib/queries";
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
  return (
    <div className="flex w-full flex-col p-12">
      <CustomerPortalForm subscription={data} />
    </div>
  );
}
