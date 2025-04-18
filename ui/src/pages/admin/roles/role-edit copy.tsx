import { LinkDto } from "@/components/landing-links";
import { MainNav } from "@/components/main-nav";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getRoleWithPermission } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router";

const rolesEditLinks: LinkDto[] = [];
export default function RoleEdit() {
  const { user } = useAuthProvider();
  const { roleId } = useParams<{ roleId: string }>();
  const {
    data: role,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: ["role", roleId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !roleId) {
        throw new Error("Missing access token or role ID");
      }
      return getRoleWithPermission(user.tokens.access_token, roleId);
    },
  });

  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;
  if (!role) return <p>Role not found</p>;

  return (
    <div className="h-full px-4 py-6 lg:px-8">
      <MainNav links={rolesEditLinks} />
      <Tabs defaultValue="general" className="h-full space-y-6">
        <TabsList>
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="permissions">Permissions</TabsTrigger>
        </TabsList>
        <TabsContent value="general"></TabsContent>
        <TabsContent value="permissions"></TabsContent>
      </Tabs>
    </div>
  );
}
