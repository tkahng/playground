import {
  Card,
  CardAction,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import ReactCountryFlag from "react-country-flag";

export type TopCountry = {
  countryName: string;
  number: number;
  country: string;
  total_reactions: number;
};
export function TopCountryCard({
  country,
  className,
  ...props
}: {
  country: TopCountry;
  className?: string;
} & React.HTMLAttributes<HTMLDivElement>) {
  return (
    <Card className={className} {...props}>
      <CardHeader>
        <CardTitle>
          #{country.number + 1} {country.countryName}{" "}
        </CardTitle>
        <CardAction className="flex items-center gap-2">
          <ReactCountryFlag
            countryCode={country.country}
            svg
            // style={{
            //   width: "1rem",
            //   height: "1rem",
            // }}
          />
        </CardAction>
      </CardHeader>
      <CardFooter>Total Reactions: {country.total_reactions}</CardFooter>
    </Card>
  );
}
