import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { settingsSidebarLinks } from "@/components/links";
import { ProfileForm } from "@/components/profile-form";
import { Separator } from "@/components/ui/separator";

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
