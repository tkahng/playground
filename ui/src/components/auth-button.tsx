import { useAuthProvider } from "@/hooks/use-auth-provider";
import AuthenticatedButton from "./authenticated-button";
import NonAuthenticatedButton from "./non-authenticated-button";

export default function AuthButton() {
  const { user } = useAuthProvider();
  return <>{user ? <AuthenticatedButton /> : <NonAuthenticatedButton />}</>;
}
