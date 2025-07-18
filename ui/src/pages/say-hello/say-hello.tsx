import { Button } from "@/components/ui/button";
import { useSSE } from "@/hooks/use-sse";
import { UserReactionsSseMessage } from "@/schema.types";
import { useEffect } from "react";
import { toast } from "sonner";

export default function SayHelloPage() {
  const { data: sseData, error: sseError } = useSSE<UserReactionsSseMessage>(
    `/api/user-reactions/sse`
  );
  //   const query = useQuery({
  //     queryKey: ["user-reactions-stats"],
  //     queryFn: async () => {
  //       const res = await fetch("/api/user-reactions");
  //       return res.json();
  //     },
  //   })
  useEffect(() => {
    if (sseError) {
      toast.error("Error", {
        description: sseError,
        action: {
          label: "Close",
          onClick: () => console.log("Close"),
        },
      });
    } else if (sseData) {
      console.log(sseData);
      toast.success("Success", {
        description: sseData.event,
        action: {
          label: "Close",
          onClick: () => console.log("Close"),
        },
      });
    }

    return () => {};
  }, [sseData, sseError]);

  return (
    <div>
      <div>SayHelloPage</div>
      <Button>Say Hello</Button>
    </div>
  );
}
