import { LinkDto } from "@/components/links";
import { RouteMap } from "@/components/route-map";
import { UserNav } from "@/components/user-nav";
const links: LinkDto[] = [
  {
    to: RouteMap.DASHBOARD,
    title: "Dashboard",
  },
  {
    to: RouteMap.SETTINGS,
    title: "Settings",
  },
];

export default function AuthenticatedButton() {
  return (
    <>
      <UserNav links={links} />
    </>
  );
}
