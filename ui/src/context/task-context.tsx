import { Task } from "@/schema.types";
import { createContext } from "react";

export const TaskContext = createContext<Task | null>(null);
