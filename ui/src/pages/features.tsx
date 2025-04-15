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
import { BarChart, Brain, CheckCircle, Globe, Lock, Zap } from "lucide-react";

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
        <Card>
          <CardHeader>
            <Zap className="h-8 w-8 text-primary mb-2" />
            <CardTitle>Lightning Fast Processing</CardTitle>
            <CardDescription>
              Experience unparalleled speed in data analysis and task
              completion.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="list-disc list-inside space-y-2 text-sm">
              <li>Process large datasets in seconds</li>
              <li>Real-time analysis and insights</li>
              <li>Optimized for high-performance computing</li>
            </ul>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <Brain className="h-8 w-8 text-primary mb-2" />
            <CardTitle>Advanced Machine Learning</CardTitle>
            <CardDescription>
              Our AI continuously learns and adapts to your specific needs.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="list-disc list-inside space-y-2 text-sm">
              <li>Personalized AI models</li>
              <li>Continuous learning from user interactions</li>
              <li>Adaptive algorithms for improved accuracy</li>
            </ul>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CheckCircle className="h-8 w-8 text-primary mb-2" />
            <CardTitle>99.9% Accuracy</CardTitle>
            <CardDescription>
              Trust in results with our industry-leading accuracy rates.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="list-disc list-inside space-y-2 text-sm">
              <li>Rigorous testing and validation</li>
              <li>Continuous model refinement</li>
              <li>Human-in-the-loop verification for critical tasks</li>
            </ul>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <BarChart className="h-8 w-8 text-primary mb-2" />
            <CardTitle>Advanced Analytics</CardTitle>
            <CardDescription>
              Gain deep insights with our powerful analytics tools.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="list-disc list-inside space-y-2 text-sm">
              <li>Customizable dashboards and reports</li>
              <li>Predictive analytics and forecasting</li>
              <li>Data visualization tools</li>
            </ul>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <Lock className="h-8 w-8 text-primary mb-2" />
            <CardTitle>Enterprise-Grade Security</CardTitle>
            <CardDescription>
              Your data is safe with our robust security measures.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="list-disc list-inside space-y-2 text-sm">
              <li>End-to-end encryption</li>
              <li>Compliance with GDPR, HIPAA, and other regulations</li>
              <li>Regular security audits and penetration testing</li>
            </ul>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <Globe className="h-8 w-8 text-primary mb-2" />
            <CardTitle>Seamless Integration</CardTitle>
            <CardDescription>
              Easily integrate NexusAI with your existing tools and workflows.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="list-disc list-inside space-y-2 text-sm">
              <li>API access for custom integrations</li>
              <li>Pre-built connectors for popular platforms</li>
              <li>Extensible plugin architecture</li>
            </ul>
          </CardContent>
        </Card>
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
