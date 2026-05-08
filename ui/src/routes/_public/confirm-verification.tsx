import ConfirmVerification from "@/pages/auth/confirm-verification";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_public/confirm-verification")({
  component: ConfirmVerification,
});
