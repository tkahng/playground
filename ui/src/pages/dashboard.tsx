import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getStats, getUserSubscriptions } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { CheckCircle, Cpu, Users, XCircle } from "lucide-react";

export default function DashboardPage() {
  const { user, checkAuth } = useAuthProvider();
  const { data, error, isError, isLoading } = useQuery({
    queryKey: ["stats"],
    queryFn: async () => {
      await checkAuth(); // Ensure user is authenticated
      if (!user) {
        throw new Error("User not found");
      }
      try {
        const stats = await getStats(user.tokens.access_token);
        const subs = await getUserSubscriptions(user.tokens.access_token);
        return {
          ...stats,
          sub: subs,
        };
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
    <div className="mx-auto px-8 py-8 justify-start items-stretch flex-1 max-w-[1200px]">
      <h1 className="mb-6 text-3xl font-bold">Overview</h1>
      <div className="grid gap-6 md:grid-cols-3 lg:grid-cols-5">
        <div className="col-span-3 gap-6 grid">
          <Card>
            <CardHeader>
              <CardTitle>Current Plan</CardTitle>
            </CardHeader>
            <CardContent className="text-4xl font-bold">
              {data.sub?.price?.product?.name || "No Plan"}
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Email Verified</CardTitle>
            </CardHeader>
            <CardContent className="text-4xl font-bold">
              {user?.user.email_verified_at ? (
                <div>
                  <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-300" />
                </div>
              ) : (
                <div>
                  <XCircle className="h-8 w-8 text-red-600 dark:text-red-300" />
                </div>
              )}
            </CardContent>
          </Card>
        </div>
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
