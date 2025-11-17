import {
  Banknote,
  Hand,
  IdCard,
  ListTodo,
  ShieldUser,
  Users,
} from "lucide-react";
import { JSX } from "react";

export default function Landing() {
  return (
    <>
      <section className="flex w-full flex-col items-center py-12 md:py-16 lg:py-24 xl:py-32">
        <div className="container px-4 md:px-6">
          <div className="flex flex-col items-center space-y-4 text-center">
            <div className="space-y-2">
              <h1 className="text-3xl font-bold tracking-tighter sm:text-2xl md:text-3xl lg:text-4xl/none">
                Welcome to my Playground
              </h1>
              <p className="mx-auto max-w-[700px] text-lg text-muted-foreground">
                This is where I experiment, learn, but most importantly have fun
                implementing cool features.
              </p>
              <p className="mx-auto max-w-[700px] text-2xl font-bold">
                If it piqued my interest, i probably will implement it here.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="flex w-full flex-col items-center py-12">
        <div className="container px-4 md:px-6">
          <h2 className="mb-12 text-center text-3xl font-bold tracking-tighter sm:text-5xl">
            What you can do here:
          </h2>
          <div className="flex flex-wrap justify-center gap-10">
            {landingFeatures.map(({ icon, title, content }, index) => (
              <div key={index} className="flex grow justify-center">
                <div className="w-90 flex flex-col items-center space-y-3 text-center">
                  {icon}
                  <h3 className="text-xl font-bold">{title}</h3>
                  <p className="text-gray-500 dark:text-gray-400">
                    {content.map((item) => {
                      return <span key={item}>{item}</span>;
                    })}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>
    </>
  );
}
export type LandingCardSectionProps = {
  title: string;
  content: string[];
  icon: JSX.Element;
};
export const landingFeatures: LandingCardSectionProps[] = [
  {
    title: "Say Hello!",
    icon: <Hand className="h-10 w-10" />,
    content: ["Whoever you are, wherever you are, you can say hello!"],
  },
  {
    title: "Teams",
    icon: <Users className="h-10 w-10" />,
    content: [
      "Create your own team with members. Easily manage member's roles and access",
    ],
  },
  {
    title: "Payment Integration",
    icon: <Banknote className="h-10 w-10" />,
    content: [
      "Teams can subscribe to different plans and manage their subscriptions.",
    ],
  },
  {
    title: "Projects & Tasks",
    icon: <ListTodo className="h-10 w-10" />,
    content: [
      "Tackle real-world problems by creating projects with tasks. Assign others and track progress.",
    ],
  },
  {
    title: "Authentication",
    icon: <IdCard className="h-10 w-10" />,
    content: ["Join your way: From Email/Password to Google, Github, etc."],
  },

  {
    title: "Admin",
    icon: <ShieldUser className="h-10 w-10" />,
    content: [
      "Manage users, roles and permissions, products, and subscriptions",
    ],
  },
];
