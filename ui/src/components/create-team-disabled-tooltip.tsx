import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { InfoIcon } from "lucide-react";

export function CreateTeamDisabledTooltip() {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <InfoIcon className="h-4 w-4" />
      </TooltipTrigger>
      <TooltipContent>
        <p>You must verifiy your email</p>
      </TooltipContent>
    </Tooltip>
  );
}
