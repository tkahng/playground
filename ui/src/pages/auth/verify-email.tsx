import { Button } from "@/components/ui/button";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useMeQuery } from "@/lib/queries";
import { useLocation } from "react-router";
export default function VerifyEmail() {
  const { search } = useLocation();
  const params = new URLSearchParams(search);
  const token = params.get("token");
  const redirectTo = params.get("redirect_to");
  const email = params.get("email");

  let navigateTo: string;
  const searchParams: URLSearchParams = new URLSearchParams();
  if (email?.length && token?.length) {
    navigateTo = `/team-invitation`;
    searchParams.set("token", token);
    searchParams.set("email", email);
  } else if (redirectTo?.length) {
    const newUrl = new URL(decodeURIComponent(redirectTo), "http://localhost");
    console.log("newUrl", newUrl);
    navigateTo = newUrl.pathname;
    for (const [key, value] of newUrl.searchParams) {
      searchParams.set(key, value);
    }
  } else {
    navigateTo = "/account/dashboard";
  }
  const { user } = useAuthProvider();

  const { isPending, isError, error } = useMeQuery();
  if (isPending) {
    return <div>Loading...</div>;
  }

  if (isError) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div className="flex w-full flex-col items-center justify-center">
      <h2>Email Confirm Success</h2>
      <p>Thank you for your verifying your email.</p>
      {user && (
        <Button asChild>
          <a href="/">Go Home</a>
        </Button>
      )}
      {!user && (
        <Button asChild>
          <a href="/signin">Sign In</a>
        </Button>
      )}
    </div>
  );
}
