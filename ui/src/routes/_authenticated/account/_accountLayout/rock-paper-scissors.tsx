import PageSectionLayout from "@/layouts/page-section";
import PlayerLayout from "@/layouts/player-layout";
import RockPaperScissors from "@/pages/account/rock-paper-scissors/rock-paper-scissors";
import { createFileRoute } from "@tanstack/react-router";

function AccountRockPaperScissors() {
  return (
    <PageSectionLayout title="Rock Paper Scissors">
      <PlayerLayout>
        <RockPaperScissors />
      </PlayerLayout>
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/account/_accountLayout/rock-paper-scissors"
)({
  component: AccountRockPaperScissors,
});
