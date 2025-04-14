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
import { Globe, Rocket, Shield, Users } from "lucide-react";

export default function LandingAboutPage() {
  return (
    <>
      <LandingTopSection
        {...{
          heading: "About NexusAI",
          description: `Empowering businesses with cutting-edge AI solutions to drive
                innovation and growth.`,
        }}
      />
      <SecondarySection>
        <Card>
          <CardHeader>
            <Rocket className="h-8 w-8 text-primary mb-2" />
            <CardTitle>Our Mission</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              To democratize AI technology and make it accessible to businesses
              of all sizes, enabling them to harness the power of artificial
              intelligence to solve complex problems and drive innovation.
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <Users className="h-8 w-8 text-primary mb-2" />
            <CardTitle>Our Team</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              We are a diverse group of AI experts, data scientists, and
              software engineers passionate about pushing the boundaries of
              what's possible with artificial intelligence.
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <Shield className="h-8 w-8 text-primary mb-2" />
            <CardTitle>Our Values</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="list-disc list-inside space-y-2">
              <li>Innovation</li>
              <li>Integrity</li>
              <li>Collaboration</li>
              <li>Customer-centricity</li>
            </ul>
          </CardContent>
        </Card>
      </SecondarySection>
      <PrimarySection>
        <h2 className="text-3xl font-bold tracking-tighter sm:text-5xl text-center mb-12">
          Our Leadership
        </h2>
        <div className="grid gap-10 sm:grid-cols-2 lg:grid-cols-3">
          <Card>
            <CardHeader>
              <img
                src="/placeholder.svg?height=100&width=100"
                alt="CEO"
                width={100}
                height={100}
                className="rounded-full mx-auto"
              />
              <CardTitle className="text-center">Jane Doe</CardTitle>
              <CardDescription className="text-center">
                CEO & Co-founder
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-center">
                With over 20 years of experience in AI and machine learning,
                Jane leads our company's vision and strategy.
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <img
                src="/placeholder.svg?height=100&width=100"
                alt="CTO"
                width={100}
                height={100}
                className="rounded-full mx-auto"
              />
              <CardTitle className="text-center">John Smith</CardTitle>
              <CardDescription className="text-center">
                CTO & Co-founder
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-center">
                John is the technical mastermind behind our AI algorithms and
                infrastructure, ensuring we stay at the cutting edge of
                technology.
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <img
                src="/placeholder.svg?height=100&width=100"
                alt="COO"
                width={100}
                height={100}
                className="rounded-full mx-auto"
              />
              <CardTitle className="text-center">Emily Chen</CardTitle>
              <CardDescription className="text-center">COO</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-center">
                Emily oversees our day-to-day operations, ensuring we deliver
                exceptional value to our customers while scaling our business.
              </p>
            </CardContent>
          </Card>
        </div>
      </PrimarySection>
      <SecondarySection cols={2}>
        {/* <section className="w-full py-12 md:py-24 lg:py-32">
        <div className="container px-4 md:px-6">
          <div className="grid gap-10 px-10 md:gap-16 lg:grid-cols-2"> */}
        <div className="space-y-4">
          <h2 className="text-3xl font-bold tracking-tighter sm:text-5xl">
            Our Story
          </h2>
          <p className="text-gray-500 dark:text-gray-400">
            Founded in 2020, NexusAI emerged from a shared vision to make AI
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
        {/* </div>
        </div>
      </section> */}
      </SecondarySection>
      <PrimarySection>
        <div className="space-y-2">
          <h2 className="text-3xl font-bold tracking-tighter sm:text-5xl">
            Join Our Team
          </h2>
          <p className="mx-auto max-w-[700px] text-gray-500 md:text-xl dark:text-gray-400">
            We're always looking for talented individuals to join our mission.
            Check out our open positions and become part of the NexusAI family.
          </p>
        </div>
        <div className="space-x-4">
          <Button size="lg">View Open Positions</Button>
          <Button variant="outline" size="lg">
            Learn About Our Culture
          </Button>
        </div>
      </PrimarySection>
    </>
  );
}
