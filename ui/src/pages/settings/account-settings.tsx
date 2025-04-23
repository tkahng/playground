// import { Button } from "@/components/ui/button";
// import {
//   Card,
//   CardContent,
//   CardDescription,
//   CardFooter,
//   CardHeader,
//   CardTitle,
// } from "@/components/ui/card";
// import { Input } from "@/components/ui/input";
// import { Label } from "@/components/ui/label";
// import { Switch } from "@/components/ui/switch";
// import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
// import { BarChart } from "lucide-react";
// import { useSearchParams } from "react-router";

import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { settingsSidebarLinks } from "@/components/links";
import { ProfileForm } from "@/components/profile-form";
import { Separator } from "@/components/ui/separator";

// enum TabType {
//   PROFILE = "profile",
//   SECURITY = "security",
//   BILLING = "billing",
//   API = "api",
// }

// export default function DashboardSubscriptions() {
//   const [searchParams, setSearchParams] = useSearchParams();
//   const getTab = (searchParams: URLSearchParams) => {
//     const tab = searchParams.get("tab");
//     if (tab === TabType.SECURITY) {
//       return TabType.SECURITY;
//     }
//     if (tab === TabType.BILLING) {
//       return TabType.BILLING;
//     }
//     if (tab === TabType.API) {
//       return TabType.API;
//     }
//     return TabType.PROFILE;
//   };
//   const onClick = (value: string) => {
//     const params = new URLSearchParams();
//     params.set("tab", value);
//     setSearchParams(params, {
//       preventScrollReset: true,
//     });
//   };
//   return (

//   );
// }

export default function AccountSettingsPage() {
  return (
    <div className="flex">
      <DashboardSidebar links={settingsSidebarLinks} />
      <div className="space-y-6 p-12">
        <div>
          <h3 className="text-lg font-medium">Profile</h3>
          <p className="text-sm text-muted-foreground">
            This is how others will see you on the site.
          </p>
        </div>
        <Separator />
        <ProfileForm />
      </div>
    </div>
  );
}
