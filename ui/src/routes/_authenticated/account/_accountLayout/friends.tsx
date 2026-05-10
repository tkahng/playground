import PageSectionLayout from "@/layouts/page-section";
import PlayerLayout from "@/layouts/player-layout";
import FriendsPage from "@/pages/account/friends/friends";
import { usePlayer } from "@/hooks/use-current-player";
import { usePlayerNotifications } from "@/hooks/use-player-notifications";
import { createFileRoute } from "@tanstack/react-router";
import { CenteredSpinner } from "@/components/centered-spinner";

function AccountFriends() {
  const { player } = usePlayer();
  usePlayerNotifications();
  if (!player) return <CenteredSpinner />;
  return (
    <PageSectionLayout title="Friends">
      <FriendsPage currentPlayerId={player.id} />
    </PageSectionLayout>
  );
}

function AccountFriendsWrapper() {
  return (
    <PlayerLayout>
      <AccountFriends />
    </PlayerLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/account/_accountLayout/friends",
)({
  component: AccountFriendsWrapper,
});
