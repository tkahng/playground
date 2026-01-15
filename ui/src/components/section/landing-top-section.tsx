import PrimarySection from "@/components/section/primary-section";

export interface LandingTopSectionProps {
  heading: string;
  description: string;
}

export default function LandingTopSection(props: LandingTopSectionProps) {
  return (
    <PrimarySection>
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tighter sm:text-5xl xl:text-6xl/none">
          {props.heading}
        </h1>
        <p className="mx-auto max-w-[700px] ">{props.description}</p>
      </div>
    </PrimarySection>
  );
}
