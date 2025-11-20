import { client } from "@/lib/client";
import { ApiError } from "@/lib/error";
import { components, operations } from "@/schema";

export const createTeam = async (
  accessToken: string,
  args: components["schemas"]["CreateTeamInput"]
) => {
  const { data, error } = await client.POST(`/api/teams`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: args,
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return data;
};

export const getUserTeamMembers = async ({
  token,
  page = 0,
  per_page = 10,
  sort_by = "team.name",
  sort_order = "asc",
  ...rest
}: {
  token: string;
} & operations["get-user-team-members"]["parameters"]["query"]) => {
  const { data, error } = await client.GET("/api/team-members", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      query: {
        page,
        per_page,
        sort_by,
        sort_order,
        ...rest,
      },
    },
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
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
export const getTeamTeamMembers = async ({
  token,
  teamId,
  page = 0,
  perPage = 10,
  search,
  active = true,
}: {
  token: string;
  teamId: string;
  page?: number;
  perPage?: number;
  search?: string;
  active?: boolean;
}) => {
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
        q: search,
        active,
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
    throw ApiError.fromErrorModel(error);
  }
  return data;
};

export const updateTeamMember = async ({
  memberId,
  token,
  role,
}: {
  token: string;
  memberId: string;
  role: components["schemas"]["TeamMember"]["role"];
}) => {
  const { data, error } = await client.PUT(
    "/api/team-members/{team-member-id}",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: { "team-member-id": memberId },
      },
      body: {
        role,
      },
    }
  );
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return data;
};
export const updateTeamMemberLastSelectedAt = async ({
  teamId,
  token,
}: {
  token: string;
  teamId: string;
}) => {
  const { data, error } = await client.PUT("/api/teams/{team-id}/select", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      path: { "team-id": teamId },
    },
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return data;
};
export const inviteTeamMember = async (
  token: string,
  teamId: string,
  body: components["schemas"]["InviteTeamMemberDto"]
) => {
  const { error } = await client.POST("/api/teams/{team-id}/invitations", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      path: { "team-id": teamId },
    },
    body,
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return true;
};
export const cancelTeamInvitation = async (
  token: string,
  teamId: string,
  invitationId: string
) => {
  const { error } = await client.DELETE(
    "/api/teams/{team-id}/invitations/{invitation-id}",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          "team-id": teamId,
          "invitation-id": invitationId,
        },
      },
    }
  );
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return true;
};
export const acceptInvitation = async (
  token: string,
  invitationToken: string
) => {
  const { error } = await client.POST("/api/team-invitations/accept", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: {
      token: invitationToken,
    },
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return true;
};

export const getUserTeamInvitations = async (
  token: string,
  page = 0,
  perPage = 10
) => {
  const { data, error } = await client.GET("/api/team-invitations", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      query: {
        page,
        per_page: perPage,
      },
    },
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return data;
};
export const getTeamMemberNotifications = async (
  token: string,
  teamMemberId: string,
  page = 0,
  perPage = 10,
  sortBy = "created_at",
  sortOrder: "asc" | "desc" | undefined = "desc"
) => {
  const { data, error } = await client.GET(
    "/api/team-members/{team-member-id}/notifications",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          "team-member-id": teamMemberId,
        },
        query: {
          page,
          per_page: perPage,
          sort_by: sortBy,
          sort_order: sortOrder,
        },
      },
    }
  );
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return data;
};
