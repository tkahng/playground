import CustomerPortalForm from "@/components/customer-portal-form";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
      {/* <div className="w-full max-w-3xl m-auto my-8 border rounded-md p border-zinc-700">
        <div className="px-5 py-4">
          <h3 className="mb-1 text-2xl font-medium">{title}</h3>
          <p className="text-zinc-300">{description}</p>
          {children}
        </div>
        {footer && (
          <div className="p-4 border-t rounded-b-md border-zinc-700 bg-zinc-900 text-zinc-500">
            {footer}
          </div>
        )}
      </div> */}
      <Card className="w-full max-w-3xl m-auto my-8 border rounded-md p border-zinc-700">
        <CardHeader>
          <CardTitle>Billing Information</CardTitle>
          <CardDescription>Billing Information</CardDescription>
        </CardHeader>
        <CardContent>
          <p>Billing Information</p>
        </CardContent>
      </Card>
      <div className="px-8 grid grid-cols-2 grid-rows-1 gap-8">
        <Card>
          <CardHeader>
            <CardTitle>Billing Information</CardTitle>
            <CardDescription>Billing Information</CardDescription>
          </CardHeader>
          <CardContent>
            <p>Billing Information</p>
          </CardContent>
          <CardFooter>
            <Button>Cancel</Button>
          </CardFooter>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Billing Information</CardTitle>
            <CardDescription>Billing Information</CardDescription>
          </CardHeader>
          <CardContent>
            <p>Billing Information</p>
          </CardContent>
          <CardFooter>
            <Button>Cancel</Button>
          </CardFooter>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Billing Information</CardTitle>
            <CardDescription>Billing Information</CardDescription>
          </CardHeader>
          <CardContent>
            <p>Billing Information</p>
          </CardContent>
          <CardFooter>
            <Button>Cancel</Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
