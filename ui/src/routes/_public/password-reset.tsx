import ConfirmPasswordReset from "@/pages/auth/confirm-password-reset";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_public/password-reset")({
  component: ConfirmPasswordReset,
});
