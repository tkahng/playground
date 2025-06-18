import { client } from "@/lib/client";
import { components, operations } from "@/schema";
import {
  RefreshTokenInput,
  SigninInput,
  SignupInput,
  UserDetailWithRoles,
  UserInfoTokens,
  UserWithAccounts,
} from "@/schema.types";

export const signIn = async (args: SigninInput): Promise<UserInfoTokens> => {
  const {
    data,
    error,
    // response: { status },
  } = await client.POST("/api/auth/signin", {
    body: {
      email: args.email,
      password: args.password,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const signOut = async (
  token: string,
  refreshToken: string
): Promise<void> => {
  const { error } = await client.POST("/api/auth/signout", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: {
      refresh_token: refreshToken,
    },
  });
  if (error) {
    throw error;
  }
};

export const refreshToken = async (
  args: RefreshTokenInput
): Promise<UserInfoTokens> => {
  const {
    data,
    error,
    // response: { status },
  } = await client.POST("/api/auth/refresh-token", {
    body: {
      refresh_token: args.refresh_token,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const signUp = async (
  args: SignupInput
): Promise<UserInfoTokens | null> => {
  const {
    data,
    error,
    // response: { status },
  } = await client.POST("/api/auth/signup", {
    body: args,
  });
  if (error) {
    throw error;
  }
  return data;
};

export const getMe = async (token: string): Promise<UserWithAccounts> => {
  const { data, error } = await client.GET("/api/auth/me", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};
export const updateMe = async (
  token: string,
  body: components["schemas"]["UpdateMeInput"]
) => {
  const { data, error } = await client.PUT("/api/auth/me", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body,
  });
  if (error) {
    throw error;
  }
  return data;
};

export const userPaginate = async (
  token: string,
  args: operations["admin-users"]["parameters"]["query"]
) => {
  const { data, error } = await client.GET("/api/admin/users", {
    params: {
      query: args,
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (error) {
    throw error;
  }
  return data;
};

export const rolesPaginate = async (
  token: string,
  args: operations["admin-roles"]["parameters"]["query"]
) => {
  const { data, error } = await client.GET("/api/admin/roles", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      query: args,
    },
  });

  if (error) {
    throw error;
  }
  return data;
};

export const getRoleWithPermission = async (token: string, id: string) => {
  const { data, error } = await client.GET(`/api/admin/roles/{id}`, {
    params: {
      query: {
        expand: ["permissions"],
      },
      path: {
        id,
      },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (error) {
    throw error;
  }
  return data;
};

export const createRole = async (
  token: string,
  body: components["schemas"]["RoleCreateInput"]
) => {
  const { data, error } = await client.POST("/api/admin/roles", {
    body,
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const updateRole = async (
  token: string,
  id: string,
  body: components["schemas"]["RoleCreateInput"]
) => {
  const { data, error } = await client.PUT(`/api/admin/roles/{id}`, {
    params: {
      path: {
        id,
      },
    },
    body,
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (error) {
    throw error;
  }
  return data;
};

export const deleteRole = async (token: string, id: string) => {
  const { data, error } = await client.DELETE(`/api/admin/roles/{id}`, {
    params: {
      path: { id },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const deleteRolePermission = async (
  token: string,
  roleId: string,
  permissionId: string
) => {
  const { data, error } = await client.DELETE(
    `/api/admin/roles/{roleId}/permissions/{permissionId}`,
    {
      params: {
        path: {
          roleId,
          permissionId,
        },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );

  if (error) {
    throw error;
  }
  return data;
};

export const createRolePermission = async (
  token: string,
  roleId: string,
  body: components["schemas"]["PermissionIdsInput"]
) => {
  const { data, error } = await client.POST(
    "/api/admin/roles/{id}/permissions",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          id: roleId,
        },
      },
      body,
    }
  );

  if (error) {
    throw error;
  }
  return data;
};

export const permissionsPaginate = async (
  token: string,
  args: operations["admin-permissions"]["parameters"]["query"]
) => {
  const { data, error } = await client.GET("/api/admin/permissions", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      query: args,
    },
  });

  if (error) {
    throw error;
  }
  return data;
};

export const deletePermission = async (token: string, id: string) => {
  const { data, error } = await client.DELETE(`/api/admin/permissions/{id}`, {
    params: {
      path: {
        id,
      },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};
export const createPermission = async (
  token: string,
  body: components["schemas"]["PermissionCreateInput"]
) => {
  const { data, error } = await client.POST("/api/admin/permissions", {
    body,
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};
export const updatePermission = async (
  token: string,
  id: string,
  body: components["schemas"]["PermissionCreateInput"]
) => {
  const { data, error } = await client.PUT(`/api/admin/permissions/{id}`, {
    params: {
      path: {
        id,
      },
    },
    body,
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (error) {
    throw error;
  }
  return data;
};

export const getUserAccounts = async (token: string, id: string) => {
  const { data, error } = await client.GET("/api/admin/user-accounts", {
    params: {
      query: {
        user_id: id,
        page: 0,
        per_page: 50,
      },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (error) {
    throw error;
  }
  return data;
};

export const getUserRoles = async (token: string, id: string) => {
  const { data, error } = await client.GET("/api/admin/roles", {
    params: {
      query: {
        page: 0,
        perPage: 50,
        user_id: id,
      },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (error) {
    throw error;
  }
  return data;
};
export const getPermission = async (token: string, id: string) => {
  const { data, error } = await client.GET(`/api/admin/permissions/{id}`, {
    params: {
      path: {
        id,
      },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (error) {
    throw error;
  }
  return data;
};
export const getUserPermissions = async (
  token: string,
  userId: string,
  reverse: boolean
) => {
  const { data, error } = await client.GET(
    "/api/admin/users/{user-id}/permissions",
    {
      params: {
        path: {
          "user-id": userId,
        },
        query: {
          page: 0,
          per_page: 50,
          reverse,
        },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  ); // TODO: add pagination
  if (error) {
    throw error;
  }
  return data;
};
export const getUserPermissions2 = async (token: string, userId: string) => {
  const { data, error } = await client.GET(
    "/api/admin/users/{user-id}/permissions",
    {
      params: {
        path: {
          "user-id": userId,
        },
        query: {
          page: 0,
          per_page: 50,
          reverse: true,
        },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  ); // TODO: add pagination
  if (error) {
    throw error;
  }
  return data;
};
export const createUser = async (
  token: string,
  body: components["schemas"]["UserCreateInput"]
) => {
  const { data, error } = await client.POST("/api/admin/users", {
    body,
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const updateUser = async (
  token: string,
  id: string,
  body: components["schemas"]["UserMutationInput"]
) => {
  const { data, error } = await client.PUT("/api/admin/users/{user-id}", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      path: {
        "user-id": id,
      },
    },
    body,
  });
  if (error) {
    throw error;
  }
  return data;
};

export const getUser = async (token: string, id: string) => {
  const { data, error } = await client.GET("/api/admin/users/{user-id}", {
    params: {
      path: {
        "user-id": id,
      },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const getUserInfo = async (
  token: string,
  id: string
): Promise<UserDetailWithRoles> => {
  const user = await getUser(token, id);
  const userRoles = await getUserRoles(token, id);
  const userPermissions = await getUserPermissions(token, id, false);
  const accoutns = await getUserAccounts(token, id);
  const userPerms: {
    created_at: string;
    description: string | null;
    id: string;
    is_directly_assigned: boolean;
    name: string;
    roles:
      | {
          created_at: string;
          description?: string;
          id: string;
          name: string;
          // permissions?: components["schemas"]["Permission"][] | null;
          updated_at: string;
        }[];
    updated_at: string;
  }[] = [];
  if (userPermissions.data?.length) {
    const ids: Set<string> = new Set();
    userPermissions.data.forEach((p) => {
      if (p.role_ids?.length) {
        p.role_ids.forEach((r) => {
          ids.add(r);
        });
      }
    });
    // const roles = await rolesPaginate(token, {
    //   page: 0,
    //   per_page: 50,
    //   ids: Array.from(ids),
    // });
    // if (roles.data?.length) {
    //   // const map = new Map(roles.data.map((x) => [x.id, x]));
    //   // roles.data.forEach((r) => {
    //   //   // const { permissions, ...rest } = r;
    //   //   // const roleList: components["schemas"]["RoleWithPermissions"][] = [];
    //   //   // if (r.permissions?.length) {
    //   //   //   r.permissions.forEach((permssion) => {
    //   //   //     const role = map.get(id.id);
    //   //   //     if (role) {
    //   //   //       roleList.push(role);
    //   //   //     }
    //   //   //   });
    //   //   // }
    //   //   // userPerms.push({
    //   //   //   ...rest,
    //   //   //   roles: roleList,
    //   //   // });
    //   // });
    // }
  }
  return {
    ...user,
    roles: userRoles.data,
    permissions: userPerms,
    accounts: accoutns.data || [],
  };
};

export const createUserRoles = async (
  token: string,
  id: string,
  body: operations["admin-create-user-roles"]["requestBody"]["content"]["application/json"]
) => {
  const { data, error } = await client.POST(
    `/api/admin/users/{user-id}/roles`,
    {
      params: {
        path: {
          "user-id": id,
        },
      },
      body,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  ); // TODO: add pagination
  if (error) {
    throw error;
  }
  return data;
};

export const removeUserRole = async (
  token: string,
  userId: string,
  roleId: string
) => {
  const { data, error } = await client.DELETE(
    `/api/admin/users/{user-id}/roles/{role-id}`,
    {
      params: {
        path: {
          "user-id": userId,
          "role-id": roleId,
        },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  ); // TODO: add pagination
  if (error) {
    throw error;
  }
  return data;
};

export const createUserPermissions = async (
  token: string,
  id: string,
  body: operations["admin-user-permissions-create"]["requestBody"]["content"]["application/json"]
) => {
  const { data, error } = await client.POST(
    `/api/admin/users/{user-id}/permissions`,
    {
      params: {
        path: {
          "user-id": id,
        },
      },
      body,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  ); // TODO: add pagination
  if (error) {
    throw error;
  }
  return data;
};

export const removeUserPermission = async (
  token: string,
  userId: string,
  permissionId: string
) => {
  const { data, error } = await client.DELETE(
    `/api/admin/users/{user-id}/permissions/{permission-id}`,
    {
      params: {
        path: {
          "user-id": userId,
          "permission-id": permissionId,
        },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  ); // TODO: add pagination
  if (error) {
    throw error;
  }
  return data;
};

export const getProductsWithPrices = async (token?: string) => {
  const { data, error } = await client.GET("/api/stripe/products", {
    headers: token
      ? {
          Authorization: `Bearer ${token}`,
        }
      : {},
  });
  if (error) {
    throw error;
  }
  if (!data) {
    throw new Error("No data");
  }

  return data;
};

export const getUserSubscriptions = async (token: string) => {
  const { data, error } = await client.GET("/api/subscriptions/active", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw error;
  }
  return data ? data : null;
};

export const getCheckoutSession = async (token: string, id: string) => {
  const { data, error } = await client.GET(
    "/api/subscriptions/checkout-session/{checkoutSessionId}",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          checkoutSessionId: id,
        },
      },
    }
  );
  if (error) {
    throw error;
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};

export const createCheckoutSession = async (
  token: string,
  { price_id }: { price_id: string }
) => {
  const { data, error } = await client.POST(
    "/api/subscriptions/checkout-session",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      body: {
        price_id,
      },
    }
  );
  if (error) {
    throw error;
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};

export const createBillingPortalSession = async (token: string) => {
  const { data, error } = await client.POST(
    "/api/subscriptions/billing-portals",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
  if (error) {
    throw error;
  }
  if (!data) {
    throw new Error("No data");
  }
  return data.url;
};

export const getAuthUrl = async ({
  provider,
  redirect,
}: {
  provider: "google" | "github";
  redirect?: string;
}) => {
  const { data, error } = await client.GET("/api/auth/authorization-url", {
    params: {
      query: {
        provider,
        redirect_to: redirect || "",
      },
    },
  });
  if (error) {
    throw error;
  }
  if (!data) {
    throw new Error("No data");
  }
  return data.url;
};

export const requestVerification = async (token: string) => {
  const { error } = await client.POST("/api/auth/request-verification", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw error;
  }
  return true;
};

export const confirmVerification = async (token: string, type: string) => {
  const { error } = await client.POST("/api/auth/verify", {
    body: {
      type,
      token,
    },
  });
  if (error) {
    throw error;
  }
};

export const taskProjectList = async (
  token: string,
  teamId: string,
  args: operations["task-project-list"]["parameters"]["query"]
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
    }
  );
  if (error) {
    throw error;
  }
  if (!data) {
    throw new Error("No data");
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
    }
  );
  if (error) {
    throw error;
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};

export const taskProjectCreate = async (
  token: string,
  teamId: string,
  args: operations["task-project-create"]["requestBody"]["content"]["application/json"]
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
    }
  );
  if (error) {
    throw error;
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};

export const taskProjectCreateWithAi = async (
  token: string,
  teamId: string,
  args: operations["task-project-create-with-ai"]["requestBody"]["content"]["application/json"]
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
    }
  );
  if (error) {
    throw error;
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};

export const taskList = async (
  token: string,
  taskProjectId: string,
  args: operations["task-list"]["parameters"]["query"]
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
    }
  );
  if (error) {
    throw error;
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};

export const createTask = async (
  token: string,
  taskProjectId: string,
  args: operations["task-project-tasks-create"]["requestBody"]["content"]["application/json"]
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
    }
  );
  if (error) {
    throw error;
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};

export const taskPositionStatus = async (
  token: string,
  taskId: string,
  args: operations["update-task-position-status"]["requestBody"]["content"]["application/json"]
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
    }
  );
  if (error) {
    throw error;
  }
  return data;
};

export const taskProjectUpdate = async (
  token: string,
  taskProjectId: string,
  args: operations["task-project-update"]["requestBody"]["content"]["application/json"]
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
    }
  );
  if (error) {
    throw error;
  }
  return data;
};

export const getStats = async (token: string) => {
  const { data, error } = await client.GET("/api/stats", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw error;
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};

export const checkPasswordReset = async (token: string) => {
  const { data, error } = await client.GET("/api/auth/check-password-reset", {
    params: {
      query: {
        token,
      },
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const confirmPasswordReset = async (
  token: string,
  password: string,
  confirmPassword: string
) => {
  const { data, error } = await client.POST(
    "/api/auth/confirm-password-reset",
    {
      body: {
        token,
        password,
        confirm_password: confirmPassword,
      },
    }
  );
  if (error) {
    throw error;
  }
  return data;
};

export const requestPasswordReset = async (email: string) => {
  const { error } = await client.POST("/api/auth/request-password-reset", {
    body: {
      email,
    },
  });
  if (error) {
    throw error;
  }
  return true;
};

export const resetPassword = async (
  token: string,
  currentPassword: string,
  newPassword: string
) => {
  const { data, error } = await client.POST("/api/auth/password-reset", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: {
      previous_password: currentPassword,
      new_password: newPassword,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const adminResetUserPassword = async (
  token: string,
  userId: string,
  newPassword: string
) => {
  const { data, error } = await client.PUT(
    "/api/admin/users/{user-id}/password",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          "user-id": userId,
        },
      },
      body: {
        password: newPassword,
      },
    }
  );
  if (error) {
    throw error;
  }
  return data;
};

export const adminStripeProducts = async (
  token: string,
  args: operations["admin-stripe-products"]["parameters"]["query"]
) => {
  const { data, error } = await client.GET("/api/admin/products", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      query: args,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const adminStripeProduct = async (token: string, id: string) => {
  const { data, error } = await client.GET("/api/admin/products/{product-id}", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      path: { "product-id": id },
      query: {
        expand: ["prices", "permissions"],
      },
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const adminStripeProductRolesCreate = async (
  token: string,
  id: string,
  body: operations["admin-create-product-permissions"]["requestBody"]["content"]["application/json"]
) => {
  const { data, error } = await client.POST(
    "/api/admin/products/{product-id}/permissions",
    {
      params: {
        path: { "product-id": id },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
      body,
    }
  );
  if (error) {
    throw error;
  }
  return data;
};

export const adminStripeProductPermissionsDelete = async (
  token: string,
  productId: string,
  permissionId: string
) => {
  const { data, error } = await client.DELETE(
    "/api/admin/products/{product-id}/permissions/{permission-id}",
    {
      params: {
        path: {
          "product-id": productId,
          "permission-id": permissionId,
        },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
  if (error) {
    throw error;
  }
  return data;
};

export const adminStripeSubscriptions = async (
  token: string,
  args: operations["admin-stripe-subscriptions"]["parameters"]["query"]
) => {
  const { data, error } = await client.GET("/api/admin/subscriptions", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      query: args,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const adminStripeSubscription = async (token: string, id: string) => {
  const { data, error } = await client.GET(
    "/api/admin/subscriptions/{subscription-id}",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: { "subscription-id": id },
        query: {
          expand: ["user", "product", "price"],
        },
      },
    }
  );
  if (error) {
    throw error;
  }
  return data;
};

export const deleteUser = async (token: string) => {
  const { data, error } = await client.DELETE("/api/auth/me", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const protectedApi = async (token: string, args: string) => {
  const { data, error } = await client.GET("/api/protected/{permission-name}", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      path: {
        "permission-name": args,
      },
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const permissionsList = async () => {
  const { data, error } = await client.GET("/api/permissions", {
    params: {
      query: {
        page: 0,
        perPage: 50,
      },
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const getUserTeams = async (token: string) => {
  const { data, error } = await client.GET("/api/teams", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      query: {
        page: 0,
        per_page: 50,
        sort_by: "name",
        sort_order: "asc",
      },
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const getTeamBySlug = async (token: string, slug: string) => {
  const { data, error } = await client.GET("/api/teams/slug/{team-slug}", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      path: { "team-slug": slug },
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const getTeamTeamMembers = async (
  token: string,
  teamId: string,
  page: number,
  perPage: number
) => {
  const { data, error } = await client.GET("/api/teams/{team-id}/members", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      path: {
        "team-id": teamId,
      },
      query: {
        page,
        per_page: perPage,
      },
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const updateTeam = async (
  token: string,
  teamId: string,
  body: components["schemas"]["UpdateTeamDto"]
) => {
  const { data, error } = await client.PUT("/api/teams/{team-id}", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      path: { "team-id": teamId },
    },
    body,
  });
  if (error) {
    throw error;
  }
  return data;
};
