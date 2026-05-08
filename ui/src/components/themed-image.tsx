import { cn } from "@/lib/utils";
export type ThemedImageSourc = {
  dark: string;
  light: string;
};

export const ThemedImage = ({
  className,
  dark,
  light,
}: {
  className?: string;
} & ThemedImageSourc) => {
  return (
    <>
      <img
        src={dark}
        className={cn("hidden dark:block w-full h-auto object-contain", className)}
      />
      <img
        src={light}
        className={cn("dark:hidden w-full h-auto object-contain", className)}
      />
    </>
  );
};
