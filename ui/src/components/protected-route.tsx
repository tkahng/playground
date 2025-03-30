import { useAuthProvider } from "@/hooks/use-auth-provider";
import { JSX } from "react";
import { Navigate } from "react-router";

export default function ProtectedRoute({
  children,
}: {
  children: JSX.Element;
}) {
  const { user } = useAuthProvider();

  return user ? children : <Navigate to="/login" />;
}
