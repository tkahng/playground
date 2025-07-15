import { Brain } from "lucide-react";
import { Link } from "react-router";

export default function PlaygroundLogo() {
  return (
    <Link className="flex items-center justify-center" to={"/"}>
      <Brain className="h-6 w-6 text-primary" />
      <span className="ml-4 text-2xl font-bold text-primary">Playground</span>
    </Link>
  );
}
