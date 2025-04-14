import { UserSubscriptionContext } from "@/context/user-subscription-context";
import { useContext } from "react";

export const useSubscription = () => {
  return useContext(UserSubscriptionContext);
};
