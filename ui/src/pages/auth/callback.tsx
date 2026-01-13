import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { decodeRedirectTo } from "@/lib/url";
import { useEffect, useRef } from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";

export default function CallbackComponent() {
  const auth = useAuthProvider();
  const navigate = useNavigate();
  const isMounted = useRef(false);
  useEffect(() => {
    if (!isMounted.current) {
      isMounted.current = true;
      const params = new URLSearchParams(window.location.search);
      const redirectTo = decodeRedirectTo(
        params.get("redirect_to"),
        RouteMap.ACCOUNT_DASHBOARD
      );
      const code = params.get("refresh_token");
      const error = params.get("error");
      console.log("OAuth callback:", { code, error });
      if (error) {
        // Handle error (e.g., display an error message)
        console.error("OAuth error:", error);
        isMounted.current = false;
        toast.error(error);
        navigate("/"); // Redirect to login page or error page
        return;
      }

      if (code) {
        // Exchange the authorization code for an access token
        // (This part depends on your OAuth2 provider and backend implementation)
        auth
          .getOrRefreshToken(code)
          .then(() => {
            // Store the access token and refresh token
            isMounted.current = false;
            // Redirect to a protected route
            navigate({
              pathname: redirectTo.pathname,
              search: redirectTo.search,
            });
          })
          .catch((error) => {
            // Handle error (e.g., display an error message)
            console.error("Error exchanging code for token:", error);
            toast.error(error);
            isMounted.current = false;
            navigate("/signin"); // Redirect to login page or error page
          });
      }
    }
  }, [auth, navigate]);

  return (
    <div>
      <h1>Processing OAuth2 Callback...</h1>
    </div>
  );
}
