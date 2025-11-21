import { Card, CardContent } from "@/components/ui/card";
import { Link } from "react-router";

export default function Dashboard() {
  return (
    <div className="flex flex-col">
      <Card className="">
        <CardContent className="m-8">
          <p>Hello 👋, welcome to your account!</p>
          <p>There are things in the works for this page, so stay tuned!</p>
          <br />
          <p>In the mean time, here are some things to do on this site:</p>
          <ul className="list-disc mx-6">
            <li>
              If you skipped the <b>Email Verification</b> step, visit the{" "}
              <Link
                to="/account/settings"
                className="text-primary hover:text-accent-foreground underline hover:no-underline"
              >
                Settings tab
              </Link>{" "}
              to resend the verification email.
            </li>
            <li>
              If you have not made a <b>Team</b>, visit the{" "}
              <Link
                to="/account/teams"
                className="text-primary hover:text-accent-foreground underline hover:no-underline"
              >
                Teams tab
              </Link>
              , create a team and invite your friends to collaborate on
              projects!
            </li>
            <li>
              Create <b>Projects</b> to manage your <b>Tasks</b> using a kanban
              board! Assign tasks to your team members, and get notified of
              deadlines!
            </li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}
