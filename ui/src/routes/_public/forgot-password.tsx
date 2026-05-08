import ResetPasswordRequestPage from "@/pages/auth/reset-password";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_public/forgot-password")({
  component: ResetPasswordRequestPage,
});
