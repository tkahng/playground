import { useNavigate, useRouterState } from "@tanstack/react-router";
import { useEffect } from "react";
import { toast } from "sonner";

export const ToastListener = ({ children }: { children: React.ReactNode }) => {
  const location = useRouterState({ select: (s) => s.location });
  const navigate = useNavigate();

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const error = params.get("error");

    if (error) {
      params.delete("error");
      navigate({
        search: () => Object.fromEntries(params.entries()),
        replace: true,
      });
      toast.error("Error", {
        description: error,
        action: {
          label: "Close",
          onClick: () => console.log("Close"),
        },
      });
    }
  }, [location, navigate]);

  return children;
};
