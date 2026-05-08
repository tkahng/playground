import { Outlet } from "@tanstack/react-router";
import { PropsWithChildren } from "react";

function TaskLayout({ children }: PropsWithChildren) {
  return (
    <div className="flex">
      <div className="flex-1 space-y-6 w-full">
        {children ?? <Outlet />}
      </div>
    </div>
  );
}

export default TaskLayout;
