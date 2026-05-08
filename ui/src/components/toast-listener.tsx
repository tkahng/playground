import { useNavigate, useRouterState } from "@tanstack/react-router";
import { useEffect } from "react";
import { toast } from "sonner";

export const ToastListener = ({ children }: { children: React.ReactNode }) => {
  const navigate = useNavigate();
  const location = useRouterState({ select: (s) => s.location });

  useEffect(() => {
    const params = new URLSearchParams(location.searchStr);
    const error = params.get("error");

    if (error) {
      params.delete("error");
      // @ts-expect-error – search schema not declared per-route; runtime is correct
      navigate({ search: () => Object.fromEntries(params.entries()), replace: true });
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
