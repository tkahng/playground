import { LinkDto } from "@/components/links";
import PlaygroundLogo from "@/components/playground-logo";
import { UserNav } from "@/components/user-nav";
import { cn } from "@/lib/utils";
import { Link, useRouterState } from "@tanstack/react-router";
import {
  Sheet,
  SheetContent,
  SheetTitle,
  SheetTrigger,
  SheetClose,
} from "./ui/sheet";
import { Menu } from "lucide-react";
import { Button } from "./ui/button";

export function PlaygroundLandingHeader({
  leftLinks,
  rightLinks,
}: {
  leftLinks?: LinkDto[];
  rightLinks?: LinkDto[];
}) {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const combinedLinks = [...(leftLinks?.length ? leftLinks : [])];
  return (
    <header>
      <nav className={cn("flex h-14 items-center box-border justify-between")}>
        <div className="hidden lg:flex grow items-center space-x-4">
          <PlaygroundLogo />
          {leftLinks?.length
            ? leftLinks?.map(({ to, title, current, badge }) => (
                <Link
                  key={to}
                  className={cn(
                    current
                      ? current(pathname)
                        ? "underline"
                        : "text-muted-foreground"
                      : to === pathname
                        ? "underline"
                        : "text-muted-foreground",
                    "flex items-center gap-1.5 text-sm font-medium underline-offset-4 hover:underline",
                  )}
                  to={to}
                >
                  {title}
                  {badge && (
                    <span className="rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-semibold text-primary-foreground leading-none">
                      {badge}
                    </span>
                  )}
                </Link>
              ))
            : null}
        </div>

        <div className="hidden lg:flex shrink items-center space-x-4">
          {rightLinks?.length && rightLinks.length > 0
            ? rightLinks.map(({ to, title, current, badge }) => (
                <Link
                  key={to}
                  className={cn(
                    current
                      ? current(pathname)
                        ? "underline"
                        : "text-muted-foreground"
                      : to === pathname
                        ? "underline"
                        : "text-muted-foreground",
                    "flex items-center gap-1.5 text-sm font-medium underline-offset-4 hover:underline",
                  )}
                  to={to}
                >
                  {title}
                  {badge && (
                    <span className="rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-semibold text-primary-foreground leading-none">
                      {badge}
                    </span>
                  )}
                </Link>
              ))
            : null}
        </div>
        <div className="lg:hidden flex grow items-center space-x-4">
          {!!combinedLinks.length && (
            <Sheet>
              <SheetTrigger asChild className="lg:hidden">
                <Button variant="ghost" size="icon">
                  <Menu className="h-6 w-6" />
                  <span className="sr-only">Toggle menu</span>
                </Button>
              </SheetTrigger>
              <SheetContent
                side="left"
                className="w-75 sm:w-100 px-4 md:px-6 lg:px-8 py-2"
              >
                <SheetTitle className="sr-only">Menu</SheetTitle>
                <div className="h-8"></div>
                <nav className="flex flex-col gap-4">
                  {[...(leftLinks?.length ? leftLinks : [])].map((item) => (
                    <SheetClose asChild key={item.to}>
                      <Link
                        key={item.to}
                        to={item.to}
                        className={cn(
                          item.current
                            ? item.current(pathname)
                              ? "underline"
                              : "text-muted-foreground"
                            : item.to === pathname
                              ? "underline"
                              : "text-muted-foreground",
                          "flex items-center gap-1.5 text-sm font-medium underline-offset-4 hover:underline",
                        )}
                      >
                        {item.title}
                        {item.badge && (
                          <span className="rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-semibold text-primary-foreground leading-none">
                            {item.badge}
                          </span>
                        )}
                      </Link>
                    </SheetClose>
                  ))}
                </nav>
              </SheetContent>
            </Sheet>
          )}
          <PlaygroundLogo />
        </div>
        <UserNav />
      </nav>
    </header>
  );
}
