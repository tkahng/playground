import CountAnimation from "@/components/count-animation";
import { UserReactionsStatsWithReactions } from "@/schema.types";

export default function StatsCard({
  stats,
}: {
  stats: UserReactionsStatsWithReactions;
}) {
  return (
    <div>
      <div>Total Reactions: {stats?.total_reactions}</div>
      <CountAnimation number={stats.total_reactions} />
    </div>
  );
}
