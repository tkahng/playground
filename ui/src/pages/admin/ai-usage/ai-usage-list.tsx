import { CenteredSpinner } from "@/components/centered-spinner";
import { DataTable } from "@/components/data-table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { adminAiUsageList } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { useState } from "react";
import { useSearchParams } from "react-router";

export default function AdminAiUsageListPage() {
  const { user } = useAuthProvider();
  const [searchParams, setSearchParams] = useSearchParams();

  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "20", 10);
  const [teamIdFilter, setTeamIdFilter] = useState(
    searchParams.get("team_id") || ""
  );

  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const next =
      typeof updater === "function"
        ? updater({ pageIndex, pageSize })
        : updater;
    setSearchParams((prev) => {
      const p = new URLSearchParams(prev);
      p.set("page", String(next.pageIndex));
      p.set("per_page", String(next.pageSize));
      return p;
    });
  };

  const applyTeamFilter = () => {
    setSearchParams((prev) => {
      const p = new URLSearchParams(prev);
      p.set("page", "0");
      if (teamIdFilter.trim()) p.set("team_id", teamIdFilter.trim());
      else p.delete("team_id");
      return p;
    });
  };

  const activeTeamId = searchParams.get("team_id") || undefined;

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["admin-ai-usage", pageIndex, pageSize, activeTeamId],
    queryFn: async () => {
      if (!user?.tokens.access_token) throw new Error("Missing access token");
      return adminAiUsageList(user.tokens.access_token, {
        page: pageIndex,
        per_page: pageSize,
        team_id: activeTeamId,
      });
    },
  });

  if (isLoading) return <CenteredSpinner />;
  if (isError) return <div>Error: {error.message}</div>;

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">
        AI token consumption records across all teams, newest first.
      </p>

      <div className="flex gap-2 max-w-sm">
        <Input
          placeholder="Filter by team ID..."
          value={teamIdFilter}
          onChange={(e) => setTeamIdFilter(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && applyTeamFilter()}
        />
        <Button variant="outline" onClick={applyTeamFilter}>
          Filter
        </Button>
        {activeTeamId && (
          <Button
            variant="ghost"
            onClick={() => {
              setTeamIdFilter("");
              setSearchParams((prev) => {
                const p = new URLSearchParams(prev);
                p.delete("team_id");
                p.set("page", "0");
                return p;
              });
            }}
          >
            Clear
          </Button>
        )}
      </div>

      <DataTable
        columns={[
          {
            accessorKey: "team_id",
            header: "Team ID",
            cell: ({ row }) => (
              <span className="font-mono text-xs">
                {row.original.team_id ?? "—"}
              </span>
            ),
          },
          {
            accessorKey: "user_id",
            header: "User ID",
            cell: ({ row }) => (
              <span className="font-mono text-xs">{row.original.user_id}</span>
            ),
          },
          {
            accessorKey: "prompt_tokens",
            header: "Prompt",
            cell: ({ row }) => row.original.prompt_tokens.toLocaleString(),
          },
          {
            accessorKey: "completion_tokens",
            header: "Completion",
            cell: ({ row }) =>
              row.original.completion_tokens.toLocaleString(),
          },
          {
            accessorKey: "total_tokens",
            header: "Total",
            cell: ({ row }) => (
              <span className="font-semibold">
                {row.original.total_tokens.toLocaleString()}
              </span>
            ),
          },
          {
            accessorKey: "created_at",
            header: "When",
            cell: ({ row }) =>
              new Date(row.original.created_at).toLocaleString(),
          },
        ]}
        data={data?.data ?? []}
        rowCount={data?.meta.total ?? 0}
        paginationState={{ pageIndex, pageSize }}
        onPaginationChange={onPaginationChange}
        paginationEnabled
      />
    </div>
  );
}
