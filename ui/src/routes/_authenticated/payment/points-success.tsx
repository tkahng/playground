import PointsPaymentSuccessPage from "@/pages/payment/points-payment-success";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/payment/points-success")({
  component: PointsPaymentSuccessPage,
});
