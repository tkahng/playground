import { Outlet } from "react-router";

function TaskLayout() {
  return (
    <div className="flex">
      {/* <div className="w-64">
        <TaskProjectSidebar />
      </div> */}
      <div className="flex-1 space-y-6 w-full">
        <Outlet />
      </div>
    </div>
  );
}

export default TaskLayout;
