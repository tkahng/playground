import { useAuthProvider } from "@/hooks/use-auth-provider";
import { taskProjectList } from "@/lib/queries";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { Link, useLocation } from "react-router";

export default function TaskProjectSidebar() {
  const { user: auth } = useAuthProvider();
  const { pathname } = useLocation();
  const {
    data: projects,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ["recent-projects"],
    select(data) {
      return data.data;
    },
    queryFn: async () => {
      if (!auth?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      const data = await taskProjectList(auth.tokens.access_token, {
        page: 1,
        per_page: 5,
        sort_by: "updated_at",
        sort_order: "desc",
      });
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
      {/* <div>
        {projects?.map((project) => (
          <Link key={project.id} to={`/projects/${project.id}`}>
            {project.name}
          </Link>
        ))}
      </div> */}
      <nav className="flex flex-col w-64 py-8 space-y-2 justify-start border-r grow-0">
        {/* {backLink && <BackLink to={backLink.to} name={backLink.title} />} */}
        {projects?.map((item) => (
          <Link
            key={item.id}
            to={`/projects/${item.id}`}
            className={cn(
              // buttonVariants({ variant: "ghost" }),
              pathname.startsWith(`/projects/${item.id}`)
                ? "underline"
                : "text-muted-foreground",
              "text-sm font-normal hover:text-primary transition-colors hover:bg-muted rounded-md p-2"
            )}
          >
            <span>{item.name}</span>
          </Link>
        ))}
      </nav>
    </>
  );
}
