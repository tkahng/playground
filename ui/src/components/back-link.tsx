import { ChevronLeft } from "lucide-react";
import { Link } from "react-router";

export default function BackLink({ name, to }: { to: string; name?: string }) {
  return (
    <Link
      to={to}
      className="flex items-center gap-2 text-sm text-muted-foreground"
    >
      <ChevronLeft className="h-4 w-4" />
      {name ?? `Back to ${name}`}
    </Link>
  );
}
