import { client } from "@/lib/client";
import { ApiError } from "@/lib/error";
import { components, operations } from "@/schema";

export const workflowList = async (
  token: string,
  teamId: string,
  args?: operations["workflow-list"]["parameters"]["query"],
) => {
  const { data, error } = await client.GET("/api/teams/{team-id}/workflows", {
    headers: { Authorization: `Bearer ${token}` },
    params: { path: { "team-id": teamId }, query: args },
  });
  if (error) throw ApiError.fromErrorModel(error);
  return data ?? [];
};

export const taskProjectList = async (
  token: string,
  teamId: string,
  args: operations["task-project-list"]["parameters"]["query"],
) => {
  const { data, error } = await client.GET(
    "/api/teams/{team-id}/task-projects",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          "team-id": teamId,
        },
        query: args,
      },
    },
  );
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  if (!data) {
    throw new ApiError("No data");
  }
  return data;
};

export const findTaskById = async (token: string, taskId: string) => {
  const { data, error } = await client.GET(`/api/tasks/{task-id}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      path: {
        "task-id": taskId,
      },
    },
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return data;
};
export const updateTaskById = async (
  token: string,
  taskId: string,
  body: components["schemas"]["UpdateTaskDto"],
) => {
  const { data, error } = await client.PUT(`/api/tasks/{task-id}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      path: {
        "task-id": taskId,
      },
    },
    body,
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return data;
};

export const taskProjectGet = async (token: string, id: string) => {
  const { data, error } = await client.GET(
    "/api/task-projects/{task-project-id}",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        query: {
          expand: ["tasks"],
        },
        path: {
          "task-project-id": id,
        },
      },
    },
  );
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};

export const taskProjectCreate = async (
  token: string,
  teamId: string,
  args: operations["task-project-create"]["requestBody"]["content"]["application/json"],
) => {
  const { data, error } = await client.POST(
    "/api/teams/{team-id}/task-projects",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          "team-id": teamId,
        },
      },
      body: args,
    },
  );
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  if (!data) {
    throw new ApiError("No data");
  }
  return data;
};

export const taskProjectCreateWithAi = async (
  token: string,
  teamId: string,
  args: operations["task-project-create-with-ai"]["requestBody"]["content"]["application/json"],
) => {
  const { data, error } = await client.POST(
    "/api/teams/{team-id}/task-projects/ai",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          "team-id": teamId,
        },
      },
      body: args,
    },
  );
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};

export const deleteTaskProject = async ({
  token,
  projectId,
}: {
  token: string;
  projectId: string;
}) => {
  const { error } = await client.DELETE(
    "/api/task-projects/{task-project-id}",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          "task-project-id": projectId,
        },
      },
    },
  );
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
};

export const taskList = async (
  token: string,
  taskProjectId: string,
  args: operations["task-list"]["parameters"]["query"],
) => {
  const { data, error } = await client.GET(
    "/api/task-projects/{task-project-id}/tasks",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          "task-project-id": taskProjectId,
        },
        query: args,
      },
    },
  );
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};

export const createTask = async (
  token: string,
  taskProjectId: string,
  args: operations["task-project-tasks-create"]["requestBody"]["content"]["application/json"],
) => {
  const { data, error } = await client.POST(
    "/api/task-projects/{task-project-id}",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          "task-project-id": taskProjectId,
        },
      },
      body: args,
    },
  );
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};
export const updateTask = async (
  token: string,
  taskId: string,
  args: operations["task-update"]["requestBody"]["content"]["application/json"],
) => {
  const { data, error } = await client.PUT("/api/tasks/{task-id}", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      path: {
        "task-id": taskId,
      },
    },
    body: args,
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return data;
};
export const updateTaskPositionStatus = async (
  token: string,
  taskId: string,
  args: operations["update-task-position-status"]["requestBody"]["content"]["application/json"],
) => {
  const { data, error } = await client.PUT(
    `/api/tasks/{task-id}/position-status`,
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          "task-id": taskId,
        },
      },
      body: args,
    },
  );
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return data;
};

export const workflowStatusCreate = async (
  token: string,
  teamId: string,
  workflowId: string,
  body: operations["workflow-status-create"]["requestBody"]["content"]["application/json"],
) => {
  const { data, error } = await client.POST(
    "/api/teams/{team-id}/workflows/{workflow-id}/statuses",
    {
      headers: { Authorization: `Bearer ${token}` },
      params: { path: { "team-id": teamId, "workflow-id": workflowId } },
      body,
    },
  );
  if (error) throw ApiError.fromErrorModel(error);
  return data;
};

export const workflowStatusUpdate = async (
  token: string,
  teamId: string,
  workflowId: string,
  statusId: string,
  body: operations["workflow-status-update"]["requestBody"]["content"]["application/json"],
) => {
  const { data, error } = await client.PUT(
    "/api/teams/{team-id}/workflows/{workflow-id}/statuses/{workflow-status-id}",
    {
      headers: { Authorization: `Bearer ${token}` },
      params: {
        path: {
          "team-id": teamId,
          "workflow-id": workflowId,
          "workflow-status-id": statusId,
        },
      },
      body,
    },
  );
  if (error) throw ApiError.fromErrorModel(error);
  return data;
};

export const workflowStatusDelete = async (
  token: string,
  teamId: string,
  workflowId: string,
  statusId: string,
) => {
  const { error } = await client.DELETE(
    "/api/teams/{team-id}/workflows/{workflow-id}/statuses/{workflow-status-id}",
    {
      headers: { Authorization: `Bearer ${token}` },
      params: {
        path: {
          "team-id": teamId,
          "workflow-id": workflowId,
          "workflow-status-id": statusId,
        },
      },
    },
  );
  if (error) throw ApiError.fromErrorModel(error);
};

export const workflowStatusReorder = async (
  token: string,
  teamId: string,
  workflowId: string,
  statusIds: string[],
) => {
  const { data, error } = await client.PUT(
    "/api/teams/{team-id}/workflows/{workflow-id}/statuses/reorder",
    {
      headers: { Authorization: `Bearer ${token}` },
      params: { path: { "team-id": teamId, "workflow-id": workflowId } },
      body: { status_ids: statusIds },
    },
  );
  if (error) throw ApiError.fromErrorModel(error);
  return data;
};

export const taskProjectUpdate = async (
  token: string,
  taskProjectId: string,
  args: operations["task-project-update"]["requestBody"]["content"]["application/json"],
) => {
  const { data, error } = await client.PUT(
    "/api/task-projects/{task-project-id}",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          "task-project-id": taskProjectId,
        },
      },
      body: args,
    },
  );
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return data;
};
