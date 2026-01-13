import { SubscriptionWithPrice } from "@/schema.types";
import { createContext } from "react";

export const UserSubscriptionContext =
  createContext<SubscriptionWithPrice | null>(null);
