import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getStats } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { Cpu, LineChart, Users } from "lucide-react";

export default function DashboardPage() {
  const { user } = useAuthProvider();
  const { data, error, isError, isLoading } = useQuery({
    queryKey: ["stats"],
    queryFn: async () => {
      if (!user) {
        throw new Error("User not found");
      }
      try {
        return await getStats(user.tokens.access_token);
      } catch (error) {
        // await checkAuth();
        throw error;
      }
    },
  });
  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error?.message}</div>;
  }
  if (!data) {
    return <div>No data</div>;
  }
  return (
    <div className="px-4 md:px-6 lg:px-8 flex-col gap-4 p-4 grow">
      <h1 className="mb-6 text-3xl font-bold">Dashboard</h1>
      <div className="grid gap-6 md:grid-cols-3 lg:grid-cols-5">
        <Card className="col-span-3">
          <CardHeader>
            <CardTitle>API Usage Over Time</CardTitle>
          </CardHeader>
          <CardContent className="pl-2">
            <LineChart className="h-[200px] w-full" />
          </CardContent>
        </Card>
        <div className="grid gap-6 md:grid-rows-1 lg:grid-rows-2 col-span-2">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Projects Done / Total
              </CardTitle>
              <Cpu className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {data?.task_stats.completed_projects} /{" "}
                {data?.task_stats.total_projects}
              </div>
              <Progress
                value={
                  (data?.task_stats.completed_projects /
                    data?.task_stats.total_projects) *
                  100
                }
                className="mt-2"
              />
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Tasks Done / Total
              </CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {data?.task_stats.completed_tasks} /{" "}
                {data?.task_stats.total_tasks}
              </div>
              <Progress
                value={
                  (data?.task_stats.completed_tasks /
                    data?.task_stats.total_tasks) *
                  100
                }
                className="mt-2"
              />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
