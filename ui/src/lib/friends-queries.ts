import { client } from "@/lib/client";
import { ApiError } from "@/lib/error";
import { components } from "@/schema";

export type Player = components["schemas"]["Player"];

export type FriendshipStatus = "pending" | "accepted" | "declined" | "blocked";

export interface Friendship {
  id: string;
  requesting_player_id: string;
  invited_player_id: string;
  status: FriendshipStatus;
  responded_at?: string | null;
  created_at: string;
  updated_at: string;
  requesting_player?: Player | null;
  invited_player?: Player | null;
}

export interface FriendshipResponse {
  data: Friendship;
}

export interface FriendshipsListResponse {
  data: Friendship[] | null;
  meta: {
    page: number;
    per_page: number;
    total: number;
    next_page?: number | null;
    prev_page?: number | null;
    has_more: boolean;
  };
}

export interface FriendshipNullableResponse {
  data: Friendship | null;
}

class FriendsQueries {
  async listFriends({
    token,
    page = 0,
    per_page = 20,
  }: {
    token: string;
    page?: number;
    per_page?: number;
  }): Promise<FriendshipsListResponse> {
    const { data, error } = await client.GET("/api/players/friends", {
      headers: { Authorization: `Bearer ${token}` },
      params: { query: { page, per_page } },
    } as Parameters<typeof client.GET>[1]);
    if (error) throw ApiError.fromErrorModel(error);
    return data as unknown as FriendshipsListResponse;
  }

  async listFriendRequests({
    token,
    page = 0,
    per_page = 20,
  }: {
    token: string;
    page?: number;
    per_page?: number;
  }): Promise<FriendshipsListResponse> {
    const { data, error } = await client.GET("/api/players/friends/requests", {
      headers: { Authorization: `Bearer ${token}` },
      params: { query: { page, per_page } },
    } as Parameters<typeof client.GET>[1]);
    if (error) throw ApiError.fromErrorModel(error);
    return data as unknown as FriendshipsListResponse;
  }

  async sendFriendRequest({
    token,
    invitedPlayerId,
  }: {
    token: string;
    invitedPlayerId: string;
  }): Promise<FriendshipResponse> {
    const { data, error } = await client.POST(
      "/api/players/friends/requests",
      {
        headers: { Authorization: `Bearer ${token}` },
        body: { invited_player_id: invitedPlayerId },
      } as Parameters<typeof client.POST>[1],
    );
    if (error) throw ApiError.fromErrorModel(error);
    return data as unknown as FriendshipResponse;
  }

  async acceptFriendRequest({
    token,
    requestId,
  }: {
    token: string;
    requestId: string;
  }): Promise<FriendshipResponse> {
    const { data, error } = await client.POST(
      "/api/players/friends/requests/{request-id}/accept",
      {
        headers: { Authorization: `Bearer ${token}` },
        params: { path: { "request-id": requestId } },
      } as Parameters<typeof client.POST>[1],
    );
    if (error) throw ApiError.fromErrorModel(error);
    return data as unknown as FriendshipResponse;
  }

  async declineFriendRequest({
    token,
    requestId,
  }: {
    token: string;
    requestId: string;
  }): Promise<FriendshipResponse> {
    const { data, error } = await client.POST(
      "/api/players/friends/requests/{request-id}/decline",
      {
        headers: { Authorization: `Bearer ${token}` },
        params: { path: { "request-id": requestId } },
      } as Parameters<typeof client.POST>[1],
    );
    if (error) throw ApiError.fromErrorModel(error);
    return data as unknown as FriendshipResponse;
  }

  async removeFriend({
    token,
    friendshipId,
  }: {
    token: string;
    friendshipId: string;
  }): Promise<void> {
    const { error } = await client.DELETE(
      "/api/players/friends/{friendship-id}",
      {
        headers: { Authorization: `Bearer ${token}` },
        params: { path: { "friendship-id": friendshipId } },
      } as Parameters<typeof client.DELETE>[1],
    );
    if (error) throw ApiError.fromErrorModel(error);
  }

  async blockPlayer({
    token,
    playerId,
  }: {
    token: string;
    playerId: string;
  }): Promise<FriendshipResponse> {
    const { data, error } = await client.POST("/api/players/block", {
      headers: { Authorization: `Bearer ${token}` },
      body: { player_id: playerId },
    } as Parameters<typeof client.POST>[1]);
    if (error) throw ApiError.fromErrorModel(error);
    return data as unknown as FriendshipResponse;
  }

  async unblockPlayer({
    token,
    playerId,
  }: {
    token: string;
    playerId: string;
  }): Promise<void> {
    const { error } = await client.DELETE("/api/players/block/{player-id}", {
      headers: { Authorization: `Bearer ${token}` },
      params: { path: { "player-id": playerId } },
    } as Parameters<typeof client.DELETE>[1]);
    if (error) throw ApiError.fromErrorModel(error);
  }

  async getFriendship({
    token,
    playerId,
  }: {
    token: string;
    playerId: string;
  }): Promise<FriendshipNullableResponse> {
    const { data, error } = await client.GET(
      "/api/players/{player-id}/friendship",
      {
        headers: { Authorization: `Bearer ${token}` },
        params: { path: { "player-id": playerId } },
      } as Parameters<typeof client.GET>[1],
    );
    if (error) throw ApiError.fromErrorModel(error);
    return data as unknown as FriendshipNullableResponse;
  }
}

export const friendsQueries = new FriendsQueries();
