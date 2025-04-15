import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getAdvancedRoute, getBasicRoute, getProRoute } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";

type Props = {
  route: "basic" | "pro" | "advanced";
};

export default function ProtectedRoutePage(props: Props) {
  const { user } = useAuthProvider();
  const { data, isPending, isError, error } = useQuery({
    queryKey: ["basic-route"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (props.route === "basic") {
        return getBasicRoute(user.tokens.access_token);
      } else if (props.route === "pro") {
        return getProRoute(user.tokens.access_token);
      } else if (props.route === "advanced") {
        return getAdvancedRoute(user.tokens.access_token);
      }
    },
  });
  if (isPending) {
    return <div>Loading...</div>;
  }

  if (isError) {
    return (
      <>
        <div>Error: {error.message}</div>
        <div>This is a protected route</div>
        <div>You need to have a {props.route} permission.</div>
        <div>Try subscribing to a correct plan.</div>
      </>
    );
  }
  return (
    <div>
      <h1>{data}</h1>
    </div>
  );
}
