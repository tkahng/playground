import { client } from "@/lib/client";
import { components, operations } from "@/schema";
import {
  AuthenticatedDTO,
  RefreshTokenInput,
  SigninInput,
  SignupInput,
  User,
  UserDetailWithRoles,
} from "@/schema.types";

export const signIn = async (
  args: SigninInput
): Promise<AuthenticatedDTO | null> => {
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
    throw new Error(error.detail);
  }
  return data || null;
};

export const refreshToken = async (
  args: RefreshTokenInput
): Promise<AuthenticatedDTO | null> => {
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
    throw new Error(error.detail);
  }
  return data;
};

export const signUp = async (
  args: SignupInput
): Promise<AuthenticatedDTO | null> => {
  const {
    data,
    error,
    // response: { status },
  } = await client.POST("/api/auth/signup", {
    body: args,
  });
  if (error) {
    throw new Error(error.detail);
  }
  return data;
};

export const getMe = async (token: string): Promise<User> => {
  const { data, error } = await client.GET("/api/auth/me", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw new Error(error.detail);
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
    throw new Error(error.detail);
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
    throw new Error(error.detail);
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
    throw new Error(error.detail);
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
    throw new Error(error.detail);
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
    throw new Error(error.detail);
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
    throw new Error(error.detail);
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
    throw new Error(error.detail);
  }
  return data;
};

export const createRolePermission = async (
  token: string,
  roleId: string,
  body: components["schemas"]["RolePermissionsUpdateInput"]
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
    throw new Error(error.detail);
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
    throw new Error(error.detail);
  }
  return data;
};

export const updatePermission = async (
  token: string,
  id: string,
  body: components["schemas"]["PermissionCreateInput"]
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
    throw new Error(error.detail);
  }
  return data;
};
export const getUserRoles = async (token: string, id: string) => {
  const { data, error } = await client.GET("/api/admin/roles", {
    params: {
      query: {
        page: 1,
        perPage: 50,
        user_id: id,
      },
    },
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (error) {
    throw new Error(error.detail);
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
    throw new Error(error.detail);
  }
  return data;
};
export const getUserPermissions = async (
  token: string,
  userId: string,
  reverse: boolean
) => {
  const { data, error } = await client.GET(
    "/api/admin/users/{userId}/permissions",
    {
      params: {
        path: {
          userId,
        },
        query: {
          page: 1,
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
    throw new Error(error.detail);
  }
  return data;
};
export const getUserPermissions2 = async (token: string, userId: string) => {
  const { data, error } = await client.GET(
    "/api/admin/users/{userId}/permissions",
    {
      params: {
        path: {
          userId,
        },
        query: {
          page: 1,
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
    throw new Error(error.detail);
  }
  return data;
};

export const getUser = async (token: string, id: string) => {
  const { data, error } = await client.GET("/api/admin/users/{id}", {
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
    throw new Error(error.detail);
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
  const userPerms: {
    created_at: string;
    description?: string;
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
    const roles = await rolesPaginate(token, {
      page: 1,
      per_page: 50,
      ids: Array.from(ids),
    });
    if (roles.data?.length) {
      const map = new Map(roles.data.map((x) => [x.id, x]));
      userPermissions.data.forEach((r) => {
        const { role_ids, ...rest } = r;
        const roleList: components["schemas"]["RoleWithPermissions"][] = [];
        if (r.role_ids?.length) {
          r.role_ids.forEach((id) => {
            const role = map.get(id);
            if (role) {
              roleList.push(role);
            }
          });
        }
        userPerms.push({
          ...rest,
          roles: roleList,
        });
      });
    }
  }
  return {
    ...user,
    roles: userRoles.data,
    permissions: userPerms,
  };
};

export const createUserRoles = async (
  token: string,
  id: string,
  body: operations["admin-create-user-roles"]["requestBody"]["content"]["application/json"]
) => {
  const { data, error } = await client.POST(`/api/admin/users/{id}/roles`, {
    params: {
      path: {
        id,
      },
    },
    body,
    headers: {
      Authorization: `Bearer ${token}`,
    },
  }); // TODO: add pagination
  if (error) {
    throw new Error(error.detail);
  }
  return data;
};

export const removeUserRole = async (
  token: string,
  userId: string,
  roleId: string
) => {
  const { data, error } = await client.DELETE(
    `/api/admin/users/{userId}/roles/{roleId}`,
    {
      params: {
        path: {
          userId,
          roleId,
        },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  ); // TODO: add pagination
  if (error) {
    throw new Error(error.detail);
  }
  return data;
};

export const createUserPermissions = async (
  token: string,
  id: string,
  body: operations["admin-user-permissions-create"]["requestBody"]["content"]["application/json"]
) => {
  const { data, error } = await client.POST(
    `/api/admin/users/{userId}/permissions`,
    {
      params: {
        path: {
          userId: id,
        },
      },
      body,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  ); // TODO: add pagination
  if (error) {
    throw new Error(error.detail);
  }
  return data;
};

export const removeUserPermission = async (
  token: string,
  userId: string,
  permissionId: string
) => {
  const { data, error } = await client.DELETE(
    `/api/admin/users/{userId}/permissions/{permissionId}`,
    {
      params: {
        path: {
          userId,
          permissionId,
        },
      },
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  ); // TODO: add pagination
  if (error) {
    throw new Error(error.detail);
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
    throw new Error(error.detail);
  }
  if (!data) {
    throw new Error("No data");
  }

  return data;
};

export const getUserSubscriptions = async (token: string) => {
  const { data, error } = await client.GET("/api/stripe/my-subscriptions", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw new Error(error.detail);
  }
  // if !data) {
  //   throw new Error("No data");
  // }
  return data;
};

export const getCheckoutSession = async (token: string, id: string) => {
  const { data, error } = await client.GET(
    "/api/stripe/checkout-session/{checkoutSessionId}",
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
    throw new Error(error.detail);
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
  const { data, error } = await client.POST("/api/stripe/checkout-session", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: {
      price_id,
    },
  });
  if (error) {
    throw new Error(error.detail);
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};

export const createBillingPortalSession = async (token: string) => {
  const { data, error } = await client.POST("/api/stripe/billing-portal", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw new Error(error.detail);
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
        redirect_to: redirect,
      },
    },
  });
  if (error) {
    throw new Error(error.detail);
  }
  if (!data) {
    throw new Error("No data");
  }
  return data.url;
};

export const confirmVerification = async (token: string, type: string) => {
  const { error } = await client.POST("/api/auth/verify", {
    body: {
      type,
      token,
    },
  });
  if (error) {
    throw new Error(error.detail);
  }
};

export const getBasicRoute = async (token: string) => {
  const { data, error } = await client.GET(
    "/api/api/protected/basic-permission",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
  if (error) {
    throw new Error(error.detail);
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};
export const getProRoute = async (token: string) => {
  const { data, error } = await client.GET(
    "/api/api/protected/pro-permission",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
  if (error) {
    throw new Error(error.detail);
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};
export const getAdvancedRoute = async (token: string) => {
  const { data, error } = await client.GET(
    "/api/api/protected/advanced-permission",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
  if (error) {
    throw new Error(error.detail);
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};
export const api = {
  protected: {
    basic: getBasicRoute,
    pro: getProRoute,
    advanced: getAdvancedRoute,
  },
};
