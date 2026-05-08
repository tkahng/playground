import Signin from "@/pages/auth/signin";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_public/signin")({
  component: Signin,
});
