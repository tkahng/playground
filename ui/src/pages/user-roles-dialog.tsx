import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import MultipleSelector, { Option } from "@/components/ui/multiple-selector";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useUserDetail } from "@/hooks/use-user-detail";
import { rolesPaginate } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";

export function DialogDemo() {
  const { user } = useAuthProvider();
  const userDetail = useUserDetail();
  const [value, setValue] = useState<Option[]>([]);

  const { data, isLoading, error } = useQuery({
    queryKey: ["user-roles-reverse", user?.tokens.access_token, userDetail?.id],
    queryFn: () => {
      if (!user?.tokens.access_token || !userDetail?.id) {
        throw new Error("Missing access token or role ID");
      }
      return rolesPaginate(user.tokens.access_token, {
        user_id: userDetail.id,
        user_reverse: true,
        page: 1,
        per_page: 50,
      });
    },
  });

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error.message}</div>;
  }

  if (!data?.data?.length) {
    return <div>User not found</div>;
  }
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline">Assign Roles</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Assign Roles</DialogTitle>
          <DialogDescription>
            Select the roles you want to assign to this user
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="w-full px-10">
            <MultipleSelector
              value={value}
              onChange={setValue}
              defaultOptions={data.data.map((role) => ({
                label: role.name,
                value: role.id,
              }))}
              placeholder="Select roles..."
              emptyIndicator={
                <p className="text-center text-lg leading-10 text-gray-600 dark:text-gray-400">
                  no results found.
                </p>
              }
            />
          </div>
        </div>
        <DialogFooter>
          <Button type="submit">Assign roles</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
