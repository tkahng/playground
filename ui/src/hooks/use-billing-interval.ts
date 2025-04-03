import {
  BillingIntervalContext,
  BillingIntervalDispatchContext,
} from "@/context/stripe-billing-interval-context";
import { useContext } from "react";

export function useBillingInterval() {
  return useContext(BillingIntervalContext);
}

export function useBillingInvervalDispatch() {
  return useContext(BillingIntervalDispatchContext);
}
