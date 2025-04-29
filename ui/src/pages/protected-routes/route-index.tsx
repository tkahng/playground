import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function ProtectedRouteIndex() {
  return (
    <div className="">
      <Card>
        <CardHeader>
          <CardTitle>Protected Route</CardTitle>
        </CardHeader>
        <CardContent>
          <p>This is a protected route</p>
          <p>You need to have a permission.</p>
          <p>Try subscribing to a correct plan.</p>
        </CardContent>
      </Card>
    </div>
  );
}
