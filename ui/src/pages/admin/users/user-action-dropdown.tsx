import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Ellipsis, Pencil } from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router";

export function UserActionDropdown({ userId }: { userId: string }) {
  // const editDialog = useDialog();
  const navigate = useNavigate();
  const [dropdownOpen, setDropdownOpen] = useState(false);
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
              navigate(`${RouteMap.ADMIN_USERS}/${userId}`);
            }}
          >
            <Button variant="ghost" size="sm">
              <Pencil className="h-4 w-4" />
              <span>Edit</span>
            </Button>
          </DropdownMenuItem>
          <DropdownMenuItem
            onSelect={() => {
              setDropdownOpen(false);
              navigate(`${RouteMap.ADMIN_USERS}/${userId}?tab=roles`);
            }}
          >
            <Button variant="ghost" size="sm">
              <Pencil className="h-4 w-4" />
              <span>Assign Roles</span>
            </Button>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
}
