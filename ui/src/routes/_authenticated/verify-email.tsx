import PublicLayout from "@/layouts/public-layout";
import VerifyEmailPage from "@/pages/auth/verify-email";
import { createFileRoute } from "@tanstack/react-router";

function VerifyEmail() {
  return (
    <PublicLayout>
      <VerifyEmailPage />
    </PublicLayout>
  );
}

export const Route = createFileRoute("/_authenticated/verify-email")({
  component: VerifyEmail,
});
