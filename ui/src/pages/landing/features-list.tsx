import {
  themedAuthenticationFeatureImage,
  ThemedImage,
  themedSayHelloFeatureImage,
} from "@/pages/landing/images";
import {
  Banknote,
  Hand,
  IdCard,
  ListTodo,
  ShieldUser,
  Users,
} from "lucide-react";
export type Feature = {
  title: string;
  shortContent: string[];
  mainContent: string[];
  icon: React.JSX.Element;
  fragment: string;
  path: string;
  detailLink?: string;
  featureImageComponent?: React.JSX.Element;
};

export const features: Feature[] = [
  {
    title: "Say Hello!",
    icon: <Hand className="h-10 w-10" />,
    shortContent: ["Whoever you are, wherever you are, you can say hello!"],
    mainContent: [
      `This is a simple feature that allows you to say hello to the world. It's a great way to test out the basic functionality of the website and get a feel for how things work.`,
    ],
    fragment: "say-hello",
    path: "/features",
    featureImageComponent: (
      <ThemedImage {...themedSayHelloFeatureImage} className="max-h-100" />
    ),
  },
  {
    title: "Authentication",
    icon: <IdCard className="h-10 w-10" />,
    shortContent: [
      "Join your way: From Email/Password to Google, Github, etc.",
    ],
    mainContent: [
      `Experience seamless and secure login with various authentication options, including email/password, Google, and GitHub. Your data is protected with the latest security protocols.`,
    ],
    fragment: "authentication",
    path: "/features",
    featureImageComponent: (
      <ThemedImage
        {...themedAuthenticationFeatureImage}
        className="max-h-100"
      />
    ),
  },
  {
    title: "Teams",
    icon: <Users className="h-10 w-10" />,
    shortContent: [
      "Create your own team with members. Easily manage member's roles and access",
    ],
    mainContent: [
      `Create and manage your own teams, invite members, and assign roles with ease. Foster collaboration and streamline your workflow by working together on projects.`,
    ],
    fragment: "teams",
    path: "/features",
  },
  {
    title: "Projects & Tasks",
    icon: <ListTodo className="h-10 w-10" />,
    shortContent: [
      "Tackle real-world problems by creating projects with tasks. Assign others and track progress.",
    ],
    fragment: "projects-and-tasks",
    path: "/features",
    mainContent: [
      `Break down your work into manageable projects and tasks. Assign tasks to team members, set deadlines, and monitor progress to ensure timely completion.`,
    ],
  },
  {
    title: "Payment Integration",
    icon: <Banknote className="h-10 w-10" />,
    shortContent: [
      "Teams can subscribe to different plans and manage their subscriptions.",
    ],
    fragment: "payment-integration",
    path: "/features",
    mainContent: [
      `Integrate with Stripe to manage subscriptions and products. Offer various plans to your teams and handle billing with confidence.`,
    ],
  },
  {
    title: "Admin",
    icon: <ShieldUser className="h-10 w-10" />,
    shortContent: [
      "Manage users, roles and permissions, products, and subscriptions",
    ],
    fragment: "admin",
    path: "/features",
    mainContent: [
      `Gain full control over your platform with robust admin features. Manage users, roles, permissions, products, and subscriptions from a centralized dashboard.`,
    ],
  },
];
