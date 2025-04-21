// react hook to set errors
import { useState } from "react";

export const useErrorStatus = () => {
  const [error, setError] = useState<string | null>(null);
};
