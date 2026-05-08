import { OnboardingChecklist } from "@/components/onboarding-checklist";
import { Card, CardContent } from "@/components/ui/card";
import { Link } from "react-router";

export default function Dashboard() {
  return (
    <div className="flex flex-col">
      <OnboardingChecklist />
      <Card>
        <CardContent className="m-8">
          <p>Hello 👋, welcome to your account!</p>
          <p>
            If you skipped the <b>Email Verification</b> step, visit the{" "}
            <Link
              to="/account/settings"
              className="text-primary hover:text-accent-foreground underline hover:no-underline"
            >
              Settings tab
            </Link>{" "}
            to resend the verification email.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
