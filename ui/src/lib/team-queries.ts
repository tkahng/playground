import { client } from "@/lib/client";
import { components } from "@/schema";

export const getTeamMembers = async (accessToken: string) => {
  const { data, error } = await client.GET("/api/team-members", {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
  if (error) {
    throw error;
  }
  return data;
};

export const checkTeamSlug = async (accessToken: string, slug: string) => {
  const { data, error } = await client.POST(`/api/teams/check-slug`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: { slug },
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
export const getTeam = async (accessToken: string, teamId: string) => {
  const { data, error } = await client.GET(`/api/teams/{team-id}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    params: {
      path: { "team-id": teamId },
    },
  });
  if (error) {
    throw error;
  }
  return data;
};
