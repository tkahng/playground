import { Link } from "react-router";
import Icon from "../assets/icon.svg";

export default function PlaygroundLogo() {
  return (
    <Link className="flex items-center justify-center" to={"/"}>
      <img src={Icon} className="h-8 w-8 rounded-sm" />
      <span className="ml-4 text-2xl font-bold text-primary">Playground</span>
    </Link>
  );
}
