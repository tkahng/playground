import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { useReducer } from "react";

import { Switch } from "@/components/ui/switch";

export default function ThemeSwitcher() {
  const { setTheme, resolvedTheme } = useTheme();
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
  const [checked, setChecked] = useReducer(
    reducer,
    resolvedTheme === "light" ? false : true
  );

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
