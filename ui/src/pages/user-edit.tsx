import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { UserDetailContext } from "@/context/user-detail-context";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useUserDetail } from "@/hooks/use-user-detail";
import { getUserInfo } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router";
import { DialogDemo } from "./user-roles-dialog";

// const formSchema = z.object({
//   name: z.string().min(2, {
//     message: "name must be at least 2 characters.",
//   }),
//   description: z
//     .string()
//     .min(2, { message: "description must be at least 2 characters." })
//     .optional(),
// });
export default function UserEdit() {
  //   const navigate = useNavigate();
  //   const queryClient = useQueryClient();
  const { user } = useAuthProvider();
  const { userId } = useParams<{ userId: string }>();
  const {
    data: userInfo,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: ["userInfo", userId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !userId) {
        throw new Error("Missing access token or role ID");
      }
      return getUserInfo(user.tokens.access_token, userId);
    },
  });

  if (loading) {
    return <div>Loading...</div>;
  }
  if (error) {
    return <div>Error: {error.message}</div>;
  }
  if (!userInfo) {
    return <div>User not found</div>;
  }
  return (
    <UserDetailContext.Provider value={userInfo}>
      <div className="flex w-full flex-col items-center justify-start">
        <div className="w-full">
          <Button asChild>
            <Link to="/dashboard/users">Back to Users</Link>
          </Button>
        </div>
        <Tabs defaultValue="profile" className="w-[ 800px]">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="profile">Account</TabsTrigger>
            <TabsTrigger value="roles">roles</TabsTrigger>
            <TabsTrigger value="permissions">permissions</TabsTrigger>
          </TabsList>
          <TabsContent value="profile">
            <Card>
              <CardHeader>
                <CardTitle>Account</CardTitle>
                <CardDescription>
                  Make changes to your account here. Click save when you're
                  done.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                <div className="space-y-1">
                  <Label htmlFor="name">email</Label>
                  <Input id="name" defaultValue={userInfo.email} />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="username">Username</Label>
                  <Input id="username" defaultValue={userInfo.name} />
                </div>
              </CardContent>
              <CardFooter>
                <Button>Save changes</Button>
              </CardFooter>
            </Card>
          </TabsContent>
          <TabsContent value="roles">
            <Card>
              <CardHeader>
                <CardTitle>Password</CardTitle>
                <CardDescription>
                  Change your password here. After saving, you'll be logged out.
                </CardDescription>
                <DialogDemo />
              </CardHeader>
              <CardContent className="space-y-2">
                <TableDemo roles={userInfo.roles || []} />
              </CardContent>
              <CardFooter>
                <Button>Save password</Button>
              </CardFooter>
            </Card>
          </TabsContent>
          <TabsContent value="permissions">
            <Card>
              <CardHeader>
                <CardTitle>Password</CardTitle>
                <CardDescription>
                  Change your password here. After saving, you'll be logged out.
                </CardDescription>
                <DialogDemo />
              </CardHeader>
              <CardContent className="space-y-2">
                <Permissions />
              </CardContent>
              <CardFooter>
                <Button>Save password</Button>
              </CardFooter>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </UserDetailContext.Provider>
  );
}

interface Roles {
  id: string;
  name: string;
  description?: string;
}
export function Permissions() {
  const user = useUserDetail();
  const permissions = user?.permissions || [];
  return (
    <Table>
      <TableCaption>A list of your recent invoices.</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">Name</TableHead>
          <TableHead>Description</TableHead>
          <TableHead>Assignment</TableHead>
          <TableHead className="text-right">Delete</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {permissions.map((perms) => (
          <TableRow key={perms.id}>
            <TableCell className="font-medium">{perms.name}</TableCell>
            <TableCell>{perms.description}</TableCell>
            <TableCell>
              {perms.is_directly_assigned && "DIRECT"},{" "}
              {perms.roles.length &&
                perms.roles.map((role) => role.name).join(", ")}
            </TableCell>
            <TableCell className="text-right">
              <Button variant="destructive">Delete</Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
      {/* <TableFooter>
        <TableRow>
          <TableCell colSpan={3}>Total</TableCell>
          <TableCell className="text-right">$2,500.00</TableCell>
        </TableRow>
      </TableFooter> */}
    </Table>
  );
}
export function TableDemo({ roles }: { roles: Roles[] }) {
  return (
    <Table>
      <TableCaption>A list of your recent invoices.</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">Name</TableHead>
          <TableHead>Description</TableHead>
          <TableHead className="text-right">Delete</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {roles.map((invoice) => (
          <TableRow key={invoice.id}>
            <TableCell className="font-medium">{invoice.name}</TableCell>
            <TableCell>{invoice.description}</TableCell>
            {/* <TableCell>{invoice.paymentMethod}</TableCell> */}
            <TableCell className="text-right">
              <Button variant="destructive">Delete</Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
      {/* <TableFooter>
        <TableRow>
          <TableCell colSpan={3}>Total</TableCell>
          <TableCell className="text-right">$2,500.00</TableCell>
        </TableRow>
      </TableFooter> */}
    </Table>
  );
}
