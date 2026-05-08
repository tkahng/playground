import SayHelloPage from "@/pages/say-hello/say-hello2";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_landing/say-hello")({
  component: SayHelloPage,
});
