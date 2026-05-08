import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { decodeRedirectTo } from "@/lib/url";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useRef } from "react";
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
        console.error("OAuth error:", error);
        isMounted.current = false;
        toast.error(error);
        navigate({ to: "/" });
        return;
      }

      if (code) {
        auth
          .getOrRefreshToken(code)
          .then(() => {
            isMounted.current = false;
            navigate({
              to: redirectTo.pathname,
              search: redirectTo.search
                ? Object.fromEntries(
                    new URLSearchParams(redirectTo.search).entries()
                  )
                : undefined,
            });
          })
          .catch((err) => {
            console.error("Error exchanging code for token:", err);
            toast.error(String(err));
            isMounted.current = false;
            navigate({ to: "/signin" });
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
