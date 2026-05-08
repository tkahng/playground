import SignupPage from "@/pages/auth/signup";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_public/signup")({
  component: SignupPage,
});
