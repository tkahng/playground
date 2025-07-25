import { Moon, Sun } from "lucide-react";
import { useReducer } from "react";
import { useTheme } from "./theme-provider";

import { Switch } from "@/components/ui/switch";

export default function ThemeSwitcher() {
  const { setTheme } = useTheme();
  const reducer = (_: boolean, action: boolean) => {
    switch (action) {
      case true:
        setTheme("dark");
        return true;
      case false:
        setTheme("light");
        return false;
    }
  };
  const [checked, setChecked] = useReducer(reducer, false);

  return (
    <div className="flex items-center space-x-3">
      <Sun className="size-4" />
      <Switch
        checked={checked}
        onCheckedChange={(value) => setChecked(value)}
        aria-label="Toggle theme"
      />
      <Moon className="size-4" />
    </div>
  );
}
