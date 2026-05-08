import RockPaperScissorsPage from "@/pages/rock-paper-scissors/rock-paper-scissors";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_landing/rock-paper-scissors")({
  component: RockPaperScissorsPage,
});
