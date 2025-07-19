import {
  Card,
  CardAction,
  CardContent,
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
  key,
  className,
}: {
  country: TopCountry;
  key: string;
  className?: string;
}) {
  return (
    <Card key={key} className={className}>
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
      <CardContent>Total Reactions: {country.total_reactions}</CardContent>
    </Card>
  );
}
