import { client } from "@/lib/client";
import { components } from "@/schema";

export const getTeamMembers = async (
  accessToken: string,
  teamId: string,
  page?: number,
  perPage?: number
) => {
  const { data, error } = await client.GET("/api/teams/{team-id}/members", {
    headers: {
      Authorization: `Bearer ${accessToken}`,
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
  perPage: number,
  search?: string
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
        q: search,
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
    throw error;
  }
  return true;
};

export const getTeamInvitations = async (
  token: string,
  teamId: string,
  page: number = 0,
  perPage: number = 10
) => {
  const { data, error } = await client.GET("/api/teams/{team-id}/invitations", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    params: {
      path: { "team-id": teamId },
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
    throw error;
  }
  return true;
};

export const verifyTeamInvitation = async (
  // token: string,
  invitationToken: string
) => {
  const { error } = await client.POST("/api/team-invitations/check", {
    // headers: {
    //   Authorization: `Bearer ${token}`,
    // },
    body: {
      token: invitationToken,
    },
  });
  if (error) {
    throw error;
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
    throw error;
  }
  return true;
};

export const declineInvitation = async (
  token: string,
  invitationToken: string
) => {
  const { error } = await client.POST("/api/team-invitations/decline", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: {
      token: invitationToken,
    },
  });
  if (error) {
    throw error;
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
    throw error;
  }
  return data;
};

export const getTeamInvitationByToken = async (invitationToken: string) => {
  const { data, error } = await client.GET(
    "/api/team-invitations/token/{token}",
    {
      // headers: {
      //   Authorization: `Bearer ${token}`,
      // },
      params: {
        path: {
          token: invitationToken,
        },
      },
    }
  );
  if (error) {
    throw error;
  }
  return data;
};

export const getTeamMemberNotifications = async (
  token: string,
  teamMemberId: string,
  page = 0,
  perPage = 10
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
          sort_by: "created_at",
          sort_order: "desc",
        },
      },
    }
  );
  if (error) {
    throw error;
  }
  return data;
};

export const readTeamMemberNotification = async (
  token: string,
  teamMemberId: string,
  notiticationId: string
) => {
  const { error } = await client.POST(
    "/api/team-members/{team-member-id}/notifications/{notification-id}/read",
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      params: {
        path: {
          "team-member-id": teamMemberId,
          "notification-id": notiticationId,
        },
      },
    }
  );
  if (error) {
    throw error;
  }
  return true;
};
