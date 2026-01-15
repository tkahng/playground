import AdminDark from "@/assets/admin-dark.png";
import AdminLight from "@/assets/admin-light.png";
import AuthenticationDark from "@/assets/authentication-dark.png";
import AuthenticationLight from "@/assets/authentication-light.png";
import PaymentDark from "@/assets/payment-dark.png";
import PaymentLight from "@/assets/payment-light.png";
import ProjectsDark from "@/assets/project-dark-2.png";
import ProjectsLight from "@/assets/project-light-2.png";
import SayHelloDark from "@/assets/say-hello-preview-dark.png";
import SayHelloLight from "@/assets/say-hello-preview-light.png";
import TeamDark from "@/assets/team-dark.png";
import TeamLight from "@/assets/team-light.png";
import ChooseYourMoveDark from "@/assets/choose-your-move-dark.png";
import ChooseYourMoveLight from "@/assets/choose-your-move-light.png";
import { ThemedImage } from "@/components/themed-image";
import {
  Banknote,
  Hand,
  IdCard,
  ListTodo,
  Scissors,
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
  detailLinkText?: string;
  featureImageComponent?: React.JSX.Element;
  landingLinkText?: string;
  landingLink: string;
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
    landingLink: "/say-hello",
    landingLinkText: "Try Saying Hello!",
    detailLink: "/say-hello",
    detailLinkText: "Try Saying Hello",
    featureImageComponent: (
      <ThemedImage
        dark={SayHelloDark}
        light={SayHelloLight}
        className="max-h-100"
      />
    ),
  },
  {
    title: "Rock Paper Scissors",
    icon: <Scissors />,
    shortContent: [
      "Challenge your friends! Send a game request with your move, they submit theirs, and see who wins.",
    ],
    mainContent: [
      `This is a fun game that allows you to challenge your friends and see who wins. It's a great way to test out the basic functionality of the website and get a feel for how things work.`,
    ],
    path: "/features",
    fragment: "rps",
    landingLink: "/rock-paper-scissors",
    detailLink: "/rock-paper-scissors",
    detailLinkText: "Play Now",
    featureImageComponent: (
      <ThemedImage
        dark={ChooseYourMoveDark}
        light={ChooseYourMoveLight}
        className="max-h-100"
      />
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
    landingLink: "/features#authentication",
    detailLink: "/signup",
    detailLinkText: "Sign Up",
    featureImageComponent: (
      <ThemedImage
        dark={AuthenticationDark}
        light={AuthenticationLight}
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
    landingLink: "/features#teams",
    featureImageComponent: (
      <ThemedImage dark={TeamDark} light={TeamLight} className="max-h-100" />
    ),
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
    landingLink: "/features#projects-and-tasks",
    featureImageComponent: (
      <ThemedImage
        dark={ProjectsDark}
        light={ProjectsLight}
        className="max-h-100"
      />
    ),
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
    landingLink: "/features#payment-integration",
    featureImageComponent: (
      <ThemedImage
        dark={PaymentDark}
        light={PaymentLight}
        className="max-h-100"
      />
    ),
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
    landingLink: "/features#admin",
    featureImageComponent: (
      <ThemedImage dark={AdminDark} light={AdminLight} className="max-h-100" />
    ),
  },
];
