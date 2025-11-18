import AuthenticationDark from "@/assets/authentication-dark.png";
import AuthenticationLight from "@/assets/authentication-light.png";
import SayHelloDark from "@/assets/say-hello-preview-dark.png";
import SayHelloLight from "@/assets/say-hello-preview-light.png";
import { cn } from "@/lib/utils";
export type ThemedImageSourc = {
  dark: string;
  light: string;
};

export const themedSayHelloFeatureImage: ThemedImageSourc = {
  dark: SayHelloDark,
  light: SayHelloLight,
};
export const themedAuthenticationFeatureImage: ThemedImageSourc = {
  dark: AuthenticationDark,
  light: AuthenticationLight,
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
      <img src={dark} className={cn("hidden dark:block", className)} />
      <img src={light} className={cn("dark:hidden", className)} />
    </>
  );
};
