import { client } from "@/lib/client";
import { ApiError } from "@/lib/error";

export class UserReactionQueries {
  createReaction = async (args?: {
    token?: string;
    coords?: { latitude: number; longitude: number };
  }) => {
    const { data, error } = await client.POST("/api/user-reactions", {
      headers: args?.token
        ? { Authorization: `Bearer ${args.token}` }
        : undefined,
      body: {
        type: "hello",
        coordinates: args?.coords,
      },
    });
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  };
  getStats = async () => {
    const { data, error } = await client.GET("/api/user-reactions/stats", {});
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  };
}

export const userReactionQueries = new UserReactionQueries();
