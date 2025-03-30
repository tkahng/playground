import { JSX } from "react";
import { Navigate } from "react-router";

export default function ProtectedRoute({
  children,
}: {
  children: JSX.Element;
}) {
  const isAuthenticated = !!localStorage.getItem("auth");

  return isAuthenticated ? children : <Navigate to="/login" />;
}
