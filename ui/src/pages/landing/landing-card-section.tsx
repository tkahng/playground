import { JSX } from "react";

export type LandingCardSectionProps = {
  title: string;
  content: string[];
  icon: JSX.Element;
};
export function LandingCardSection({
  title,
  content,
  icon,
}: LandingCardSectionProps) {
  return (
    <div className="flex flex-col items-center space-y-3 text-center">
      {icon}
      <h3 className="text-xl font-bold">{title}</h3>
      <p className="text-gray-500 dark:text-gray-400">
        {content.map((item) => {
          return <p>{item}</p>;
        })}
      </p>
    </div>
  );
}
