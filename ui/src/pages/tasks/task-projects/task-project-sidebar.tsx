import { Separator } from "@/components/ui/separator";
import { useTeamMemberContext } from "@/context/team-members-context";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { taskProjectList } from "@/lib/queries";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { Link, useLocation } from "react-router";
export default function TaskProjectSidebar() {
  const { user: auth, checkAuth } = useAuthProvider();
  const { currentMember } = useTeamMemberContext();
  const { pathname } = useLocation();
  const {
    data: projects,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ["recent-projects"],
    select(data) {
      return data;
    },
    queryFn: async () => {
      await checkAuth(); // Ensure user is authenticated
      if (!auth?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      if (!currentMember?.team_id) {
        throw new Error("Current team member team ID is required");
      }
      const data = await taskProjectList(
        auth.tokens.access_token,
        currentMember.team_id,
        {
          page: 0,
          per_page: 5,
          sort_by: "updated_at",
          sort_order: "desc",
        }
      );
      if (!data?.data) {
        throw new Error("No projects found");
      }
      return data;
    },
  });
  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error?.message}</div>;
  }
  return (
    <>
      <nav className="flex flex-col w-64 space-y-2 justify-start border-r grow-0">
        <div className="flex flex-col pr-4">
          <h3 className="text-md font-medium">Recent Projects</h3>
        </div>
        <div className="flex flex-col pr-4">
          <Separator />
        </div>
        {projects?.data?.map((item) => (
          <Link
            key={item.id}
            to={`/projects/${item.id}`}
            className={cn(
              pathname.startsWith(`/projects/${item.id}`)
                ? "underline"
                : "text-muted-foreground",
              "text-md font-normal hover:text-primary transition-colors hover:bg-muted rounded-md p-2"
            )}
          >
            <span>{item.name}</span>
          </Link>
        ))}
        <div className="flex flex-col pr-4">
          <Separator />
        </div>
        {projects?.meta.has_more && (
          <Link
            to={`/projects`}
            className={cn(
              "text-sm font-normal hover:text-primary transition-colors hover:bg-muted rounded-md p-2 flex"
            )}
          >
            View all projects
          </Link>
        )}
      </nav>
    </>
  );
}
