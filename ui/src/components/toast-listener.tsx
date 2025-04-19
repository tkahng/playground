import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router";
import { toast } from "sonner";

export const ToastListener = () => {
  const location = useLocation();
  const navigate = useNavigate();

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const error = params.get("error");

    if (error) {
      toast.error(error);

      // Remove "error" from the query params
      params.delete("error");

      // Replace the URL with the updated query string
      navigate(
        {
          pathname: location.pathname,
          search: params.toString(),
        },
        { replace: true }
      );
    }
  }, [location, navigate]);

  return null; // this component doesn’t render anything
};
