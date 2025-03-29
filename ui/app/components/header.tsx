import type { JSX } from "react";
import NexusAILogo from "app/components/nexus-logo";
import { NavLink } from "app/components/nav-link";



export type LinkProps = {
    title?: string;
    icon?: JSX.Element;
    href: string;
  };
  
export function NexusAILandingHeader({
    leftLinks,
    rightLinks,
  }: {
    leftLinks?: LinkProps[];
    rightLinks?: LinkProps[];
  }) {
    return (
      <header>
        <nav className="flex h-14 items-center justify-between px-4 shadow-sm lg:px-6">
          <div className="mr-auto flex flex-grow items-center space-x-4">
            <NexusAILogo />
            {leftLinks?.length &&
              leftLinks.map(({ href, title }) => (
                <NavLink key={title} title={title} to={href} />
              ))}
          </div>
          <div className="flex shrink items-center space-x-4">
            {/* {rightLinks?.length &&
              rightLinks.map(({ href, title }) => (
                <NavLink key={title} title={title} href={href} />
              ))} */}
            {/* <ModeToggle />
            <AuthButton /> */}
          </div>
        </nav>
      </header>
    );
  }