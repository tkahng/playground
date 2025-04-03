import type { BillingIntervals } from "@/schema.types";
import type { Dispatch } from "react";
import { createContext } from "react";

export const BillingIntervalContext = createContext<BillingIntervals>("month");

export const BillingIntervalDispatchContext = createContext<
  Dispatch<BillingIntervalsAction>
>((action) => action);

export interface BillingIntervalsAction {
  interval: BillingIntervals;
}

export function tasksReducer(action: BillingIntervalsAction) {
  return action.interval;
}
