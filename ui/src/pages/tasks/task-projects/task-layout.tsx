import { Outlet } from "react-router";
import TaskProjectSidebar from "./task-project-sidebar";

function TaskLayout() {
  return (
    <div className="flex">
      <div className="w-64">
        <TaskProjectSidebar />
      </div>
      <div className="flex-1">
        <Outlet />
      </div>
    </div>
  );
}

export default TaskLayout;
