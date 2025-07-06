import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

type TeamStatsCardProps = {
  title: string;
  icon: React.ReactNode;
  value: number | string;
  description?: string;
};
export function TeamStatCard(props: TeamStatsCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{props.title}</CardTitle>
        {props.icon}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{props.value}</div>
        <p className="text-xs text-muted-foreground">{props.description}</p>
      </CardContent>
    </Card>
  );
}
