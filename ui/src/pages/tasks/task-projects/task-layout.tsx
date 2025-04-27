import { Outlet } from "react-router";
import TaskProjectSidebar from "./task-project-sidebar";

function TaskLayout() {
  return (
    <div className="flex">
      <div className="w-64">
        <TaskProjectSidebar />
      </div>
      <div className="flex-1 space-y-6 p-12 w-full">
        <Outlet />
      </div>
    </div>
  );
}

export default TaskLayout;
