import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Ellipsis, Pencil } from "lucide-react";
import { useState } from "react";

export function TeamMemberActionDropdown({ memberId }: { memberId: string }) {
  const [dropdownOpen, setDropdownOpen] = useState(false);
  if (!memberId) return null;
  return (
    <>
      <DropdownMenu open={dropdownOpen} onOpenChange={setDropdownOpen}>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon">
            <Ellipsis className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem
            onSelect={() => {
              setDropdownOpen(false);
              // navigate(`${RouteMap.ADMIN_USERS}/${memberId}?tab=roles`);
            }}
          >
            <Button variant="ghost" size="sm">
              <Pencil className="h-4 w-4" />
              <span>Cancel Invitation</span>
            </Button>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
}
