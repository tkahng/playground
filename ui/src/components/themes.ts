import { LucideProps, Moon, Settings, Sun } from "lucide-react";

export const themeMap = {
  dark: { name: "Dark", value: "dark", icon: Moon },
  light: { name: "Light", value: "light", icon: Sun },
  system: { name: "System", value: "system", icon: Settings },
} as const;

export const themes = Object.entries(themeMap).map(([, value]) => ({
  name: value.name,
  value: value.value,
  icon: value.icon,
}));

export interface ThemeData {
  name: string;
  value: Themes;
  icon: React.ForwardRefExoticComponent<
    Omit<LucideProps, "ref"> & React.RefAttributes<SVGSVGElement>
  >;
}

export type Themes = keyof typeof themeMap;
