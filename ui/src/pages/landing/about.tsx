import LandingTopSection from "@/components/section/landing-top-section";
import PrimarySection from "@/components/section/primary-section";
import SecondarySection from "@/components/section/secondary-section";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Rocket, Shield, Users } from "lucide-react";

export default function LandingAboutPage() {
  return (
    <>
      <LandingTopSection
        {...{
          heading: "About Playground",
          description: `A place of learning and experimentation.`,
        }}
      />
      <SecondarySection>
        <Card>
          <CardHeader>
            <Rocket className="h-8 w-8 text-primary mb-2" />
            <CardTitle>What this place is for</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              This website is a place for learning and experimentation. Anything
              I find interesting or useful, I will incorporate into this
              website.
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <Users className="h-8 w-8 text-primary mb-2" />
            <CardTitle>Who am I</CardTitle>
          </CardHeader>
          <CardContent>
            <p>Just a lonely developer finding his way through the world.</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <Shield className="h-8 w-8 text-primary mb-2" />
            <CardTitle>The website's tech stack</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="list-disc list-inside space-y-2">
              <li>Golang</li>
              <li>PostgreSQL</li>
              <li>React</li>
              <li>Tailwind</li>
              <li>Motion</li>
            </ul>
          </CardContent>
        </Card>
      </SecondarySection>
      <PrimarySection>
        <h2 className="text-3xl font-bold tracking-tighter sm:text-5xl text-center mb-12">
          Credits
        </h2>
        shapes by LAFS from{" "}
        <a
          href="https://thenounproject.com/browse/icons/term/shapes/"
          target="_blank"
          title="shapes Icons"
        >
          Noun Project
        </a>{" "}
        (CC BY 3.0)
      </PrimarySection>
    </>
  );
}
