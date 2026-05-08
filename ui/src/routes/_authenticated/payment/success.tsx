import PaymentSuccessPage from "@/pages/payment/payment-success";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/payment/success")({
  component: PaymentSuccessPage,
});
