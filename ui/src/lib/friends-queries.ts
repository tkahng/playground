import { client } from "@/lib/client";
import { ApiError } from "@/lib/error";
import { components, operations } from "@/schema";

export type Friendship = components["schemas"]["Friendship"];
export type Player = components["schemas"]["Player"];
export type FriendshipStatus = Friendship["status"];

type PaginatedFriendships = components["schemas"]["ApiPaginatedResponseFriendship"];
type SingleFriendship = components["schemas"]["ApiSingleResponseFriendship"];

class FriendsQueries {
  async listFriends({
    token,
    page = 0,
    per_page = 20,
  }: {
    token: string;
    page?: number;
    per_page?: number;
  }): Promise<PaginatedFriendships> {
    const { data, error } = await client.GET("/api/players/friends", {
      headers: { Authorization: `Bearer ${token}` },
      params: { query: { page, per_page } },
    });
    if (error) throw ApiError.fromErrorModel(error);
    return data;
  }

  async listFriendRequests({
    token,
    page = 0,
    per_page = 20,
  }: {
    token: string;
    page?: number;
    per_page?: number;
  }): Promise<PaginatedFriendships> {
    const { data, error } = await client.GET("/api/players/friends/requests", {
      headers: { Authorization: `Bearer ${token}` },
      params: { query: { page, per_page } },
    });
    if (error) throw ApiError.fromErrorModel(error);
    return data;
  }

  async sendFriendRequest({
    token,
    invitedPlayerId,
  }: {
    token: string;
    invitedPlayerId: string;
  }): Promise<SingleFriendship> {
    const body: components["schemas"]["SendFriendRequestBody"] = {
      invited_player_id: invitedPlayerId,
    };
    const { data, error } = await client.POST("/api/players/friends/requests", {
      headers: { Authorization: `Bearer ${token}` },
      body,
    });
    if (error) throw ApiError.fromErrorModel(error);
    return data;
  }

  async acceptFriendRequest({
    token,
    requestId,
  }: {
    token: string;
    requestId: string;
  }): Promise<SingleFriendship> {
    const { data, error } = await client.POST(
      "/api/players/friends/requests/{request-id}/accept",
      {
        headers: { Authorization: `Bearer ${token}` },
        params: { path: { "request-id": requestId } },
      },
    );
    if (error) throw ApiError.fromErrorModel(error);
    return data;
  }

  async declineFriendRequest({
    token,
    requestId,
  }: {
    token: string;
    requestId: string;
  }): Promise<SingleFriendship> {
    const { data, error } = await client.POST(
      "/api/players/friends/requests/{request-id}/decline",
      {
        headers: { Authorization: `Bearer ${token}` },
        params: { path: { "request-id": requestId } },
      },
    );
    if (error) throw ApiError.fromErrorModel(error);
    return data;
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
      },
    );
    if (error) throw ApiError.fromErrorModel(error);
  }

  async blockPlayer({
    token,
    playerId,
  }: {
    token: string;
    playerId: string;
  }): Promise<SingleFriendship> {
    const body: components["schemas"]["BlockPlayerBody"] = {
      player_id: playerId,
    };
    const { data, error } = await client.POST("/api/players/block", {
      headers: { Authorization: `Bearer ${token}` },
      body,
    });
    if (error) throw ApiError.fromErrorModel(error);
    return data;
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
    });
    if (error) throw ApiError.fromErrorModel(error);
  }

  async getFriendship({
    token,
    playerId,
  }: {
    token: string;
    playerId: string;
  }): Promise<SingleFriendship> {
    const { data, error } = await client.GET(
      "/api/players/{player-id}/friendship",
      {
        headers: { Authorization: `Bearer ${token}` },
        params: { path: { "player-id": playerId } },
      },
    );
    if (error) throw ApiError.fromErrorModel(error);
    return data;
  }

  async issuePlayerSseTicket({
    token,
  }: {
    token: string;
  }): Promise<components["schemas"]["IssuePlayerSSETicketResponseBody"]> {
    const { data, error } = await client.POST("/api/players/sse/ticket", {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (error) throw ApiError.fromErrorModel(error);
    return data;
  }
}

export const friendsQueries = new FriendsQueries();

// Re-export operation parameter types for convenience.
export type ListFriendsParams =
  operations["list-friends"]["parameters"]["query"];
export type ListFriendRequestsParams =
  operations["list-friend-requests"]["parameters"]["query"];
