import { useAuthProvider } from "@/hooks/use-auth-provider";
import AuthenticatedButton from "@/components/authenticated-button";
import NonAuthenticatedButton from "@/components/non-authenticated-button";

export default function AuthButton() {
  const { user } = useAuthProvider();
  return <>{user ? <AuthenticatedButton /> : <NonAuthenticatedButton />}</>;
}
