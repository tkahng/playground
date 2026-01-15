import { TaskContext } from "@/context/task-context";
import { useContext } from "react";

export const useTask = () => useContext(TaskContext);
