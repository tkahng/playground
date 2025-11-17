import { LandingFeatureCardProps } from "@/pages/landing/landing";
import {
  Banknote,
  Hand,
  IdCard,
  ListTodo,
  ShieldUser,
  Users,
} from "lucide-react";

export const landingFeatures: LandingFeatureCardProps[] = [
  {
    title: "Say Hello!",
    icon: <Hand className="h-10 w-10" />,
    content: ["Whoever you are, wherever you are, you can say hello!"],
    featureLink: "/features#say-hello",
  },
  {
    title: "Teams",
    icon: <Users className="h-10 w-10" />,
    content: [
      "Create your own team with members. Easily manage member's roles and access",
    ],
    featureLink: "/features#teams",
  },
  {
    title: "Payment Integration",
    icon: <Banknote className="h-10 w-10" />,
    content: [
      "Teams can subscribe to different plans and manage their subscriptions.",
    ],
    featureLink: "/features#payment-integration",
  },
  {
    title: "Projects & Tasks",
    icon: <ListTodo className="h-10 w-10" />,
    content: [
      "Tackle real-world problems by creating projects with tasks. Assign others and track progress.",
    ],
    featureLink: "/features#projects-and-tasks",
  },
  {
    title: "Authentication",
    icon: <IdCard className="h-10 w-10" />,
    content: ["Join your way: From Email/Password to Google, Github, etc."],
    featureLink: "/features#authentication",
  },

  {
    title: "Admin",
    icon: <ShieldUser className="h-10 w-10" />,
    content: [
      "Manage users, roles and permissions, products, and subscriptions",
    ],
    featureLink: "/features#admin",
  },
];
