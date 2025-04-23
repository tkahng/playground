import { useAuthProvider } from "@/hooks/use-auth-provider";
import { GetError } from "@/lib/get-erro";
import { api } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { DashboardSidebar } from "./dashboard-sidebar";
import { protectedSidebarLinks } from "./links";
type Props = {
  route: "basic" | "pro" | "advanced";
};

export default function ProtectedRoutePage(props: Props) {
  const { user } = useAuthProvider();
  const { data, isPending, isError, error } = useQuery({
    queryKey: ["protected-route"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (props.route === "basic") {
        return api.protected.basic(user.tokens.access_token);
      } else if (props.route === "pro") {
        return api.protected.pro(user.tokens.access_token);
      } else if (props.route === "advanced") {
        return api.protected.advanced(user.tokens.access_token);
      }
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
        <div className="flex">
          <DashboardSidebar links={protectedSidebarLinks} />
          <div className="flex-1 space-y-6 p-12 w-full">
            <div>
              <div>Error: {err.detail}</div>
              <div>This is a protected route</div>
              <div>You need to have a {props.route} permission.</div>
              <div>Try subscribing to a correct plan.</div>
              <div>
                {err.errors?.map((e) => (
                  <div key={e.location}>{e.message}</div>
                ))}
              </div>
            </div>
          </div>
        </div>
      );
    }
  }
  return (
    <div className="flex">
      <DashboardSidebar links={protectedSidebarLinks} />
      <div className="flex-1 space-y-6 p-12 w-full">
        <h1>{data}</h1>
      </div>
    </div>
  );
}
