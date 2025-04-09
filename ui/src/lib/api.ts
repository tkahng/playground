import { components, operations } from "@/schema";
import {
  AuthenticatedDTO,
  RefreshTokenInput,
  SigninInput,
  SignupInput,
  User,
} from "@/schema.types";
import { client } from "./client";

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

export const getRole = async (token: string, id: string) => {
  const { data, error } = await client.GET(`/api/admin/roles/{id}`, {
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

export const updateRolePermissions = async (
  token: string,
  id: string,
  body: components["schemas"]["RolePermissionsUpdateInput"]
) => {
  const { data, error } = await client.PUT(
    `/api/admin/roles/{id}/permissions`,
    {
      params: {
        path: {
          id,
        },
      },
      body,
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

export const getUserPermissions = async (token: string, id: string) => {
  const { data, error } = await client.GET(
    "/api/admin/users/{id}/permissions",
    {
      params: {
        query: {
          page: 1,
          perPage: 50,
        },
        path: {
          id,
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

export const getUserInfo = async (token: string, id: string) => {
  const user = await getUser(token, id);
  const userRoles = await getUserRoles(token, id);
  const userPermissions = await getUserPermissions(token, id);
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
          permissions?: components["schemas"]["Permission"][] | null;
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
