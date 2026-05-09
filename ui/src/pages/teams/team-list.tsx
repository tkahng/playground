import { useSearchParams } from "@/hooks/use-search-params";
import { CenteredSpinner } from "@/components/centered-spinner";
import { CreateTeamDialog } from "@/components/create-team-dialog";
import { DataTable } from "@/components/data-table";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getUserTeamMembers } from "@/lib/team-queries";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { Link } from "@tanstack/react-router";

export default function TeamListPage() {
  const { user } = useAuthProvider();
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);
  const q = searchParams.get("q") || "";
  const sortBy = searchParams.get("sort_by") || "team.name";
  const sortOrder = (searchParams.get("sort_order") || "asc") as "asc" | "desc";

  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const newState =
      typeof updater === "function"
        ? updater({ pageIndex, pageSize })
        : updater;
    setSearchParams({
      page: String(newState.pageIndex),
      per_page: String(newState.pageSize),
      q,
      sort_by: sortBy,
      sort_order: sortOrder,
    });
  };

  const { data, error, isError, isLoading } = useQuery({
    queryKey: ["get-user-team-members", user?.user.id, pageIndex, pageSize, q, sortBy, sortOrder],
    queryFn: async () => {
      if (!user) {
        throw new Error("User not found");
      }
      return getUserTeamMembers({
        token: user.tokens.access_token,
        page: pageIndex,
        per_page: pageSize,
        q: q || undefined,
        sort_by: sortBy as "team.name" | "team.created_at" | "team.updated_at" | "user.email" | "user.name" | "user.created_at" | "user.updated_at" | "last_selected_at",
        sort_order: sortOrder,
      });
    },
  });

  if (isLoading) {
    return <CenteredSpinner />;
  }
  if (isError) {
    return <div>Error: {error?.message}</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <p>Create and manage Teams for your applications.</p>
        <CreateTeamDialog />
      </div>
      <div className="flex items-center gap-2">
        <Input
          placeholder="Search teams..."
          value={q}
          onChange={(e) =>
            setSearchParams({
              page: "0",
              per_page: String(pageSize),
              q: e.target.value,
              sort_by: sortBy,
              sort_order: sortOrder,
            })
          }
          className="max-w-sm"
        />
        <Select
          value={sortBy}
          onValueChange={(v) =>
            setSearchParams({
              page: "0",
              per_page: String(pageSize),
              q,
              sort_by: v,
              sort_order: sortOrder,
            })
          }
        >
          <SelectTrigger className="w-[140px]">
            <SelectValue placeholder="Sort by" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="team.name">Name</SelectItem>
            <SelectItem value="team.created_at">Created</SelectItem>
            <SelectItem value="last_selected_at">Last visited</SelectItem>
          </SelectContent>
        </Select>
        <Select
          value={sortOrder}
          onValueChange={(v) =>
            setSearchParams({
              page: "0",
              per_page: String(pageSize),
              q,
              sort_by: sortBy,
              sort_order: v,
            })
          }
        >
          <SelectTrigger className="w-[130px]">
            <SelectValue placeholder="Order" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="asc">Ascending</SelectItem>
            <SelectItem value="desc">Descending</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <DataTable
        columns={[
          {
            accessorKey: "name",
            header: "Name",
            cell: ({ row }) => {
              return (
                <Link
                  to="/teams/$teamSlug/dashboard"
                  params={{ teamSlug: row.original.team?.slug ?? "" }}
                  className="hover:underline text-blue-500"
                >
                  {row.original.team?.name}
                </Link>
              );
            },
          },
          {
            accessorKey: "role",
            header: "Member Role",
            cell: ({ row }) => {
              return <span className="text-gray-500">{row.original.role}</span>;
            },
          },
        ]}
        data={data?.data || []}
        rowCount={data?.meta.total || 0}
        paginationState={{ pageIndex, pageSize }}
        onPaginationChange={onPaginationChange}
        paginationEnabled
      />
    </div>
  );
}
