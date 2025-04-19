import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect } from "react";
import { useNavigate } from "react-router";

export default function CallbackComponent() {
  const auth = useAuthProvider();
  const navigate = useNavigate();

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const code = params.get("refresh_token");
    const error = params.get("error");
    console.log("OAuth callback:", { code, error });
    if (error) {
      // Handle error (e.g., display an error message)
      console.error("OAuth error:", error);
      navigate("/login"); // Redirect to login page or error page
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
          // Redirect to a protected route
          navigate("/dashboard");
        })
        .catch((error) => {
          // Handle error (e.g., display an error message)
          console.error("Error exchanging code for token:", error);
          navigate("/login"); // Redirect to login page or error page
        });
    }
  }, [navigate]);

  return (
    <div>
      <h1>Processing OAuth2 Callback...</h1>
    </div>
  );
}
