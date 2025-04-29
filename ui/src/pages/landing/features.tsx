import LandingTopSection from "@/components/section/landing-top-section";
import PrimarySection from "@/components/section/primary-section";
import SecondarySection from "@/components/section/secondary-section";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Zap } from "lucide-react";

export default function Features() {
  return (
    <>
      <LandingTopSection
        {...{
          heading: "Powerful Features to Supercharge Your Workflow",
          description: `Discover how NexusAI can revolutionize your work with
                cutting-edge AI technology.`,
        }}
      />
      <SecondarySection>
        <FeatureCard
          title="AI Task Generation"
          description="Describe your project in plain language, and our AI will break it down
          into actionable tasks with deadlines."
          items={[
            "Suggested timelines and dependencies based on task complexity",
            "Automatically adjusts timelines as tasks are completed",
          ]}
        />
        <FeatureCard
          title="User Management and Security"
          description="Manage your users with ease"
          items={[
            "Customizable user roles and permissions",
            "Secure user access with authentication and authorization",
            "Product based access control",
          ]}
        />
        <FeatureCard
          title="Product and subscription integration"
          description="With Stripe Integration you can connect your products to your subscription plan."
          items={[
            "Stripe Webhook Integration",
            "Stripe Subscription Management",
            "Stripe Billing Portal Integration",
          ]}
        />
      </SecondarySection>
      <PrimarySection>
        <div className="space-y-2">
          <h2 className="text-3xl font-bold tracking-tighter sm:text-5xl">
            Ready to Get Started?
          </h2>
          <p className="mx-auto max-w-[700px] ">
            Join thousands of satisfied users and experience the power of
            NexusAI today.
          </p>
        </div>
        <div className="space-x-4">
          <Button size="lg">Start Free Trial</Button>
          <Button variant="outline" size="lg">
            Contact Sales
          </Button>
        </div>
      </PrimarySection>
    </>
  );
}
function FeatureCard(props: {
  title: string;
  description: string;
  items: string[];
}) {
  return (
    <Card>
      <CardHeader>
        <Zap className="h-8 w-8 text-primary mb-2" />
        <CardTitle className="text-2xl font-bold">{props.title}</CardTitle>
        <CardDescription className="text-lg">
          {props.description}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ul className="list-disc list-inside space-y-2">
          {props.items.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </CardContent>
    </Card>
  );
}
