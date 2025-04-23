import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { protectedSidebarLinks } from "@/components/links";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function ProtectedRouteIndex() {
  return (
    <div className="flex">
      <DashboardSidebar links={protectedSidebarLinks} />
      <div className="flex-1 space-y-6 p-12 w-full">
        <Card>
          <CardHeader>
            <CardTitle>Protected Route</CardTitle>
          </CardHeader>
          <CardContent>
            <p>This is a protected route</p>
            <p>You need to have a basic permission.</p>
            <p>Try subscribing to a correct plan.</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
