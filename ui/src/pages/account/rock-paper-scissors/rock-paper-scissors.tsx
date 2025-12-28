import { DataTable } from "@/components/data-table";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useItemDialog } from "@/hooks/use-item-dialg";
import { rpsGameQueries } from "@/lib/rps-game-queries";
import { CreateGameDialog } from "@/pages/account/rock-paper-scissors/create-game-dialog";
import { RpsGameWithParticipants } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { useSearchParams } from "react-router";

export default function RockPaperScissors() {
  const { isOpen, selectedItemId, openItem, closeDialog } = useItemDialog();

  const userInfo = useAuthProvider();

  // const [game, setGame] = useState<RpsGameWithParticipants | null>(null);
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);
  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const newState =
      typeof updater === "function"
        ? updater({ pageIndex, pageSize })
        : updater;
    setSearchParams({
      page: String(newState.pageIndex),
      per_page: String(newState.pageSize),
    });
  };
  const { data, isLoading, isError, error } = useQuery({
    queryKey: [{ key: "rps-games", page: pageIndex, per_page: pageSize }],
    queryFn: async () => {
      if (!userInfo.user?.tokens.access_token) {
        throw new Error("No access token");
      }
      return rpsGameQueries.getRpsGames({
        token: userInfo.user.tokens.access_token,
        page: pageIndex,
        per_page: pageSize,
        sort_order: "desc",
        sort_by: "created_at",
      });
    },
  });
  const selectedItem =
    data?.data?.find((i) => i.rps_game.id === selectedItemId) ?? null;

  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div>
      <h1>Rock Paper Scissors</h1>
      <div className="flex items-center justify-between">
        <p>
          Create and manage permissions for your applications. Permissions and
          can be assigned to Users.
        </p>
        <CreateGameDialog />
      </div>
      {isOpen && selectedItem && (
        <ItemDialog item={selectedItem} onClose={closeDialog} />
      )}
      <DataTable
        columns={[
          {
            header: "Result",
            cell: ({ row }) => {
              if (row.original.rps_game.status !== "completed") {
                return "pending";
              }
              if (
                row.original.invited_participant.player?.email ===
                userInfo.user?.user.email
              ) {
                return row.original.requesting_participant?.result;
              }
              return row.original.invited_participant?.result;
            },
          },
          {
            header: "Player",
            cell: ({ row }) => {
              if (
                row.original.invited_participant.player?.email ===
                userInfo.user?.user.email
              ) {
                return row.original.requesting_participant?.player?.email || "";
              }
              return row.original.invited_participant?.player?.email || "";
            },
          },
          {
            header: "Created At",
            cell: ({ row }) => {
              return new Date(
                row.original.rps_game.created_at
              ).toLocaleDateString();
            },
          },
        ]}
        onClick={(row) => {
          openItem(row.original.rps_game.id);
        }}
        data={data?.data || []}
        rowCount={data?.meta.total || 0}
        paginationState={{ pageIndex, pageSize }}
        onPaginationChange={onPaginationChange}
        paginationEnabled
      />
    </div>
  );
}

function ItemDialog({
  item,
  onClose,
}: {
  item: RpsGameWithParticipants;
  onClose: () => void;
}) {
  return (
    <div className="backdrop" onClick={onClose}>
      <div className="dialog" onClick={(e) => e.stopPropagation()}>
        <p>{item.rps_game.status}</p>
        <button onClick={onClose}>Close</button>
      </div>
    </div>
  );
}
