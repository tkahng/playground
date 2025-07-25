import { useAuthProvider } from "@/hooks/use-auth-provider";
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
          .then((data) => {
            // Store the access token and refresh token
            console.log("Token data:", data);
            isMounted.current = false;
            // Redirect to a protected route
            navigate("/");
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
