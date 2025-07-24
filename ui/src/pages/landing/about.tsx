import LandingTopSection from "@/components/section/landing-top-section";
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

      {/* <SecondarySection cols={2}>
        <div className="space-y-4">
          <h2 className="text-3xl font-bold tracking-tighter sm:text-5xl">
            Our Story
          </h2>
          <p className="text-gray-500 dark:text-gray-400">
            Founded in 2020, Playground emerged from a shared vision to make AI
            accessible and impactful for businesses worldwide. What started as a
            small team of passionate AI enthusiasts has grown into a leading
            force in the AI industry, serving clients across various sectors and
            continents.
          </p>
          <p className="text-gray-500 dark:text-gray-400">
            Our journey has been marked by continuous innovation, overcoming
            challenges, and celebrating successes alongside our clients. Today,
            we're proud to be at the forefront of the AI revolution, helping
            businesses transform their operations and unlock new possibilities.
          </p>
        </div>
        <div className="space-y-4">
          <h2 className="text-3xl font-bold tracking-tighter sm:text-5xl">
            Our Impact
          </h2>
          <ul className="grid gap-4">
            <li className="flex items-center gap-4">
              <Globe className="h-8 w-8 text-primary" />
              <div>
                <h3 className="font-bold">Global Reach</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Serving clients in over 50 countries
                </p>
              </div>
            </li>
            <li className="flex items-center gap-4">
              <Users className="h-8 w-8 text-primary" />
              <div>
                <h3 className="font-bold">Growing Community</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Over 1 million users and counting
                </p>
              </div>
            </li>
            <li className="flex items-center gap-4">
              <Rocket className="h-8 w-8 text-primary" />
              <div>
                <h3 className="font-bold">Driving Innovation</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  100+ patents filed in AI technology
                </p>
              </div>
            </li>
          </ul>
        </div>
      </SecondarySection> */}
    </>
  );
}
