import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router";
import { toast } from "sonner";

export const ToastListener = ({ children }: { children: React.ReactNode }) => {
  const location = useLocation();
  const navigate = useNavigate();

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const error = params.get("error");

    if (error) {
      params.delete("error");
      navigate(
        {
          pathname: location.pathname,
          search: params.toString(),
        },
        { replace: true }
      );
      toast.error(error);
      toast.error("Error", {
        description: error,
        action: {
          label: "Close",
          onClick: () => console.log("Close"),
        },
      });

      // Remove "error" from the query params

      // Replace the URL with the updated query string
    }
  }, [location, navigate]);
  // toast.error("Error");

  return children; // this component doesn’t render anything
};
