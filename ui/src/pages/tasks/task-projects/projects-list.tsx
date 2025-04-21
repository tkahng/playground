import { RouteMap } from "@/components/route-map";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { taskProjectList } from "@/lib/queries";
import { useInfiniteQuery } from "@tanstack/react-query";
import React from "react";
import { Link, useSearchParams } from "react-router";
import { CreateProjectAiDialog } from "./create-project-ai-dialog";
import { CreateProjectDialog } from "./create-project-dialog";

export default function ProjectListPage() {
  const { user } = useAuthProvider();
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);
  console.log(pageIndex, pageSize);
  // const queryClient = useQueryClient();
  // const onPaginationChange = (updater: Updater<PaginationState>) => {
  //   const newState =
  //     typeof updater === "function"
  //       ? updater({ pageIndex, pageSize })
  //       : updater;
  //   if (newState.pageIndex !== pageIndex || newState.pageSize !== pageSize) {
  //     setSearchParams({
  //       page: String(newState.pageIndex),
  //       per_page: String(newState.pageSize),
  //     });
  //   }
  // };

  const {
    data,
    error,
    fetchNextPage,
    hasNextPage,
    isFetching,
    isFetchingNextPage,
    status,
  } = useInfiniteQuery({
    queryKey: ["projects-list"],
    initialPageParam: 0,
    queryFn: async ({ pageParam = 10 }) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      const data = await taskProjectList(user.tokens.access_token, {
        page: pageParam,
        per_page: pageSize,
      });

      return data;
    },
    getNextPageParam: (lastPage, _allPages, lastPageParam) => {
      if (lastPage.data?.length === 0) {
        return undefined;
      }
      return lastPageParam + 1;
    },
    getPreviousPageParam: (_firstPage, _allPages, firstPageParam) => {
      if (firstPageParam <= 1) {
        return undefined;
      }
      return firstPageParam - 1;
    },
  });
  if (status === "pending") {
    return <div>Loading...</div>;
  }
  if (status === "error") {
    return <div>Error: {error?.message}</div>;
  }
  // const projects = data?.data || [];
  // const rowCount = data?.meta.total || 0;

  return (
    // <div className="flex w-full flex-col items-center justify-center">
    <div className="mx-auto w-full max-w-[1200px] py-12 px-4 @lg:px-6 @xl:px-12 @2xl:px-20 @3xl:px-24">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Projects</h1>
        <CreateProjectDialog />
        <CreateProjectAiDialog />
      </div>
      <p>
        Create and manage Projects for your applications. Projects contain
        collections of Tasks and can be assigned to Users.
      </p>

      <div className="gap-4 grid md:grid-cols-2 lg:grid-cols-3">
        {data.pages.map(
          (group, i) =>
            group?.data && (
              <React.Fragment key={i}>
                {group.data.map((project) => (
                  <div key={project.id}>
                    <Link to={`${RouteMap.TASK_PROJECTS}/${project.id}`}>
                      <Card className="col-span-1">
                        <CardHeader>
                          <CardTitle>{project.name}</CardTitle>
                          <CardDescription>
                            {project.description}
                          </CardDescription>
                        </CardHeader>
                      </Card>
                    </Link>
                  </div>
                ))}
              </React.Fragment>
            )
        )}
        <div>
          <button
            onClick={() => fetchNextPage()}
            disabled={!hasNextPage || isFetchingNextPage}
          >
            {isFetchingNextPage
              ? "Loading more..."
              : hasNextPage
              ? "Load More"
              : "Nothing more to load"}
          </button>
        </div>
        <div>{isFetching && !isFetchingNextPage ? "Fetching..." : null}</div>
      </div>
      {/* <DataTable
        columns={[
          {
            accessorKey: "name",
            header: "Name",
            cell: ({ row }) => {
              return (
                <NavLink
                  to={`${RouteMap.TASK_PROJECTS}/${row.original.id}`}
                  className="hover:underline text-blue-500"
                >
                  {row.original.name}
                </NavLink>
              );
            },
          },
          {
            accessorKey: "description",
            header: "Description",
          },
          // {
          //   id: "actions",
          //   cell: ({ row }) => {
          //     return (
          //       <div className="flex flex-row gap-2 justify-end">
          //         <RoleEllipsisDropdown
          //           roleId={row.original.id}
          //           onDelete={(roleId) => {
          //             // mutation.mutate(roleId);
          //           }}
          //         />
          //       </div>
          //     );
          //   },
          // },
        ]}
        data={projects}
        rowCount={rowCount}
        paginationState={{ pageIndex, pageSize }}
        onPaginationChange={onPaginationChange}
        paginationEnabled
      /> */}
    </div>
  );
}
