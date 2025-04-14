import { ModeToggle } from "./mode-toggle";
import NexusAILogo from "./nexus-logo";

export default function NexusAIAuthHeader() {
  return (
    <header className="flex h-14 items-center justify-self-stretch px-4 shadow-sm lg:px-6">
      <NexusAILogo />

      <ModeToggle />
    </header>
  );
}
