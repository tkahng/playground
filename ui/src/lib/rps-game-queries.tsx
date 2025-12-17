import { client } from "@/lib/client";
import { ApiError } from "@/lib/error";
import { components } from "@/schema";

export class RpsGameQueries {
  async PutUserPlayer({
    token,
    displayName,
  }: {
    token: string;
    displayName: string;
  }) {
    const { data, error } = await client.PUT(`/api/players/me`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      body: {
        display_name: displayName,
      },
    });
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  }
  async getUserPlayer({ token }: { token: string }) {
    const { data, error } = await client.GET(`/api/players/me`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  }
  async getRpsGameWithToken({ token }: { token: string }) {
    const { data, error } = await client.POST(`/api/games/rps/invites/verify`, {
      body: {
        token,
      },
    });
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  }
  async submitMoveWithToken({
    token,
    status,
    move,
  }: components["schemas"]["SubmitMoveWithTokenInput"]) {
    const { data, error } = await client.POST(
      `/api/games/rps/token/submit-move`,
      {
        body: {
          token,
          move,
          status,
        },
      }
    );
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  }
}

export const rpsGameQueries = new RpsGameQueries();
