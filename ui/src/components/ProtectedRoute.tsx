import { AuthContext } from "@/context/auth-context";
import { JSX, useContext } from "react";
import { Navigate } from "react-router";

const ProtectedRoute = ({
  children,
}: {
  children: JSX.Element;
}): JSX.Element => {
  const { checkAuth } = useContext(AuthContext);
  const res = checkAuth();
  return res ? (children as JSX.Element) : <Navigate to="/login/signin" />;
};

export default ProtectedRoute;
