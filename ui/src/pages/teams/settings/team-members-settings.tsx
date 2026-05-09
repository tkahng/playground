import { useSearchParams } from "@/hooks/use-search-params";
import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { DataTable } from "@/components/data-table";
import { teamSettingLinks } from "@/components/links";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { getTeamTeamMembers } from "@/lib/team-queries";
import { MemberRowDropdownMenuDialog } from "@/pages/teams/settings/member-row-dropdown-dialog";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { CheckCircle, XCircle } from "lucide-react";
import { InviteTeamMemberDialog } from "./invite-team-member-dialog";
import { CenteredSpinner } from "@/components/centered-spinner";

type Role = "owner" | "member" | "guest";

export default function TeamMembersSettingPage() {
  const { user } = useAuthProvider();
  const { team } = useTeam();
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);
  const q = searchParams.get("q") || "";
  const sortBy = searchParams.get("sort_by") || "user.name";
  const sortOrder = (searchParams.get("sort_order") || "asc") as "asc" | "desc";
  const roleFilter = (searchParams.get("role") || "all") as Role | "all";

  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const newState =
      typeof updater === "function"
        ? updater({ pageIndex, pageSize })
        : updater;
    if (newState.pageIndex !== pageIndex || newState.pageSize !== pageSize) {
      setSearchParams({
        page: String(newState.pageIndex),
        per_page: String(newState.pageSize),
        q,
        sort_by: sortBy,
        sort_order: sortOrder,
        role: roleFilter,
      });
    }
  };

  const roles: Role[] | undefined =
    roleFilter !== "all" ? [roleFilter] : undefined;

  const { data, isPending, isError, error } = useQuery({
    queryKey: [
      "team-team-members",
      team?.id,
      pageIndex,
      pageSize,
      q,
      sortBy,
      sortOrder,
      roleFilter,
    ],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (!team?.id) {
        throw new Error("Current team member team ID is required");
      }
      return getTeamTeamMembers({
        token: user.tokens.access_token,
        teamId: team.id,
        page: pageIndex,
        perPage: pageSize,
        search: q || undefined,
        active: true,
        sortBy,
        sortOrder,
        roles,
      });
    },
  });

  if (isPending) {
    return <CenteredSpinner />;
  }

  if (isError) {
    return <div>Error: {error.message}</div>;
  }
  if (!team) {
    return <div>Team not found</div>;
  }

  return (
    <div className="flex">
      <DashboardSidebar links={teamSettingLinks(team?.slug)} />
      <div className="flex-1 space-y-6 p-12 w-full">
        <div className="flex items-center justify-between">
          <p>
            Manage your team's members. Invite team members to join your team.
          </p>
          <InviteTeamMemberDialog />
        </div>
        <div className="flex items-center gap-2">
          <Input
            placeholder="Search members..."
            value={q}
            onChange={(e) =>
              setSearchParams({
                page: "0",
                per_page: String(pageSize),
                q: e.target.value,
                sort_by: sortBy,
                sort_order: sortOrder,
                role: roleFilter,
              })
            }
            className="max-w-sm"
          />
          <Select
            value={roleFilter}
            onValueChange={(v) =>
              setSearchParams({
                page: "0",
                per_page: String(pageSize),
                q,
                sort_by: sortBy,
                sort_order: sortOrder,
                role: v,
              })
            }
          >
            <SelectTrigger className="w-[130px]">
              <SelectValue placeholder="Role" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All roles</SelectItem>
              <SelectItem value="owner">Owner</SelectItem>
              <SelectItem value="member">Member</SelectItem>
              <SelectItem value="guest">Guest</SelectItem>
            </SelectContent>
          </Select>
          <Select
            value={sortBy}
            onValueChange={(v) =>
              setSearchParams({
                page: "0",
                per_page: String(pageSize),
                q,
                sort_by: v,
                sort_order: sortOrder,
                role: roleFilter,
              })
            }
          >
            <SelectTrigger className="w-[140px]">
              <SelectValue placeholder="Sort by" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="user.name">Name</SelectItem>
              <SelectItem value="user.email">Email</SelectItem>
              <SelectItem value="user.created_at">Joined</SelectItem>
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
                role: roleFilter,
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
          data={data.data || []}
          rowCount={data.meta.total || 0}
          paginationState={{ pageIndex, pageSize }}
          onPaginationChange={onPaginationChange}
          paginationEnabled
          columns={[
            {
              header: "Name",
              accessorKey: "user.name",
            },
            {
              header: "Email",
              accessorKey: "user.email",
            },
            {
              header: "Role",
              accessorKey: "role",
            },
            {
              header: "Billing Access",
              accessorKey: "has_billing_access",
              cell: ({ row }) => {
                return row.original.has_billing_access ? (
                  <CheckCircle className="text-green-600 dark:text-green-300" />
                ) : (
                  <XCircle className="text-muted-foreground" />
                );
              },
            },
            {
              id: "actions",
              cell: ({ row }) => {
                return (
                  <div className="flex flex-row gap-2 justify-end">
                    <MemberRowDropdownMenuDialog member={row.original} />
                  </div>
                );
              },
            },
          ]}
        />
      </div>
    </div>
  );
}
