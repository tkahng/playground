import { useAuthProvider } from "@/hooks/use-auth-provider";
import { protectedApi } from "@/lib/api";
import { GetError } from "@/lib/get-error";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router";

export default function ProtectedRoutePage() {
  const { user } = useAuthProvider();
  const { permission } = useParams<{ permission: string }>();
  console.log("permission", permission);
  const { data, isPending, isError, error } = useQuery({
    queryKey: ["protected-route", permission],
    queryFn: async () => {
      if (!user?.tokens.access_token || !permission) {
        throw new Error("Missing access token");
      }
      return protectedApi(user.tokens.access_token, permission);
    },
    retry: false,
  });
  if (isPending) {
    return <div>Loading...</div>;
  }

  if (isError) {
    const err = GetError(error);
    console.log("err", err);
    if (err) {
      return (
        <div>
          <div>Error: {err.detail}</div>
          <div>This is a protected route</div>
          <div>You need to have a {permission} permission.</div>
          <div>Try subscribing to a correct plan.</div>
          <div>
            {err.errors?.map((e) => (
              <div key={e.location}>{e.message}</div>
            ))}
          </div>
        </div>
      );
    }
  }
  return (
    <>
      <h1>{data}</h1>
      <div>
        You can see this page because you have a {permission} permission
      </div>
    </>
  );
}
