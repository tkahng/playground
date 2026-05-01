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
  badge?: string;
};

// Ordered by discovery flow: try it first → sign up → team → project → plan → game → admin
export const features: Feature[] = [
  {
    title: "Say Hello!",
    badge: "Start here — no sign-up needed",
    icon: <Hand className="h-10 w-10" />,
    shortContent: ["Whoever you are, wherever you are, you can say hello!"],
    mainContent: [
      "Wave to the world with a single click. Your location is added to a live counter, and you can watch hellos stream in from around the globe in real time.",
    ],
    fragment: "say-hello",
    path: "/features",
    landingLink: "/say-hello",
    landingLinkText: "Try Saying Hello!",
    detailLink: "/say-hello",
    detailLinkText: "Try it now",
    featureImageComponent: (
      <ThemedImage
        dark={SayHelloDark}
        light={SayHelloLight}
        className="max-h-100"
      />
    ),
  },
  {
    title: "Authentication",
    badge: "Required for the rest",
    icon: <IdCard className="h-10 w-10" />,
    shortContent: [
      "Join your way: Email/Password, Google, GitHub, and more.",
    ],
    mainContent: [
      "Sign up with email/password or OAuth providers like Google and GitHub. Email verification, password reset, and session management all included.",
    ],
    fragment: "authentication",
    path: "/features",
    landingLink: "/features#authentication",
    detailLink: "/signup",
    detailLinkText: "Create an account",
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
    badge: "Requires account",
    icon: <Users className="h-10 w-10" />,
    shortContent: [
      "Create a team, invite members, and manage roles and access.",
    ],
    mainContent: [
      "Create and manage teams, invite members by email, and assign roles with fine-grained access control. Each team gets its own dashboard and project space.",
    ],
    fragment: "teams",
    path: "/features",
    landingLink: "/features#teams",
    detailLink: "/account/teams",
    detailLinkText: "Go to Teams",
    featureImageComponent: (
      <ThemedImage dark={TeamDark} light={TeamLight} className="max-h-100" />
    ),
  },
  {
    title: "Projects & Tasks",
    badge: "Requires a team",
    icon: <ListTodo className="h-10 w-10" />,
    shortContent: [
      "Create projects with Kanban boards. Assign tasks, set statuses, track progress.",
    ],
    fragment: "projects-and-tasks",
    path: "/features",
    mainContent: [
      "Break work into projects and tasks on a Kanban board. Assign tasks to team members, drag between columns, and monitor progress — all within your team's workspace.",
    ],
    landingLink: "/features#projects-and-tasks",
    detailLink: "/account/teams",
    detailLinkText: "Open your team",
    featureImageComponent: (
      <ThemedImage
        dark={ProjectsDark}
        light={ProjectsLight}
        className="max-h-100"
      />
    ),
  },
  {
    title: "Plans & Billing",
    badge: "Stripe-powered",
    icon: <Banknote className="h-10 w-10" />,
    shortContent: [
      "Subscribe your team to a plan and unlock protected features.",
    ],
    fragment: "payment-integration",
    path: "/features",
    mainContent: [
      "Stripe-powered subscription management. Teams can subscribe to Basic, Pro, or Advanced plans to unlock protected routes. Manage billing, view invoices, and cancel anytime.",
    ],
    landingLink: "/features#payment-integration",
    detailLink: "/pricing",
    detailLinkText: "See plans",
    featureImageComponent: (
      <ThemedImage
        dark={PaymentDark}
        light={PaymentLight}
        className="max-h-100"
      />
    ),
  },
  {
    title: "Rock Paper Scissors",
    badge: "Play with anyone",
    icon: <Scissors className="h-10 w-10" />,
    shortContent: [
      "Challenge anyone via a shareable link. Place point bets and see who wins.",
    ],
    mainContent: [
      "Create a game, pick your move, and share a link. Your opponent picks theirs and the result is revealed. Bet points on the outcome — points are earned and tracked per account.",
    ],
    path: "/features",
    fragment: "rps",
    landingLink: "/rock-paper-scissors",
    detailLink: "/rock-paper-scissors",
    detailLinkText: "Play now",
    featureImageComponent: (
      <ThemedImage
        dark={ChooseYourMoveDark}
        light={ChooseYourMoveLight}
        className="max-h-100"
      />
    ),
  },
  {
    title: "Admin",
    badge: "Admin role required",
    icon: <ShieldUser className="h-10 w-10" />,
    shortContent: [
      "Manage users, roles, permissions, products, and background jobs.",
    ],
    fragment: "admin",
    path: "/features",
    mainContent: [
      "Full admin dashboard for managing users, assigning roles and permissions, configuring Stripe products, viewing subscriptions, and monitoring background jobs.",
    ],
    landingLink: "/features#admin",
    featureImageComponent: (
      <ThemedImage dark={AdminDark} light={AdminLight} className="max-h-100" />
    ),
  },
];
