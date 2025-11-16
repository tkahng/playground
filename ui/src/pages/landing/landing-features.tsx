import { LandingCardSectionProps } from "@/pages/landing/landing-card-section";
import { Hand, IdCard, Users } from "lucide-react";

export const landingFeatures: LandingCardSectionProps[] = [
  {
    title: "Say Hello!",
    content: ["Whoever you are, wherever you are, you can say hello!"],
    icon: <Hand className="h-10 w-10" />,
  },
  {
    title: "Authentication",
    content: [
      "Signup howevery you want: from the classic email and password to social login with Google, Github, etc.Manage them in your account settings.",
    ],
    icon: <IdCard className="h-10 w-10 text-primary" />,
  },
  {
    title: "Teams",
    content: [
      "You can create teams with your favorite(or not) people!",
      "Invite new members through email, manage existing member's roles, and remove members.",
    ],
    icon: <Users className="h-10 w-10 text-primary" />,
  },

  {
    title: "Projects",
    content: [
      "Signup howevery you want: from the classic email and password to social login with Google, Github, etc.Manage them in your account settings.",
    ],
    icon: <IdCard className="h-10 w-10 text-primary" />,
  },
];
