import { client } from "@/lib/client";
import { ApiError } from "@/lib/error";
import { components, operations } from "@/schema";


export type ChallengeHouseResult = components["schemas"]["ChallengeHouseResponse"];

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
      },
    );
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  }
  async submitMoveToGame({
    token,
    gameId,
    ...rest
  }: components["schemas"]["SubmitMoveToGameInput"] & {
    token: string;
    gameId: string;
  }) {
    const { data, error } = await client.POST(
      `/api/games/rps/{game-id}/submit-move`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: rest,
        params: {
          path: {
            "game-id": gameId,
          },
        },
      },
    );
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  }

  async getRpsGames({
    token,
    ...rest
  }: {
    token: string;
  } & operations["find-current-players-rps-games"]["parameters"]["query"]) {
    const { data, error } = await client.GET(
      `/api/players/current-player/games/rps`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
        params: {
          query: rest,
        },
      },
    );
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  }
  async getRpsGame({
    token,
    ...rest
  }: {
    token: string;
  } & { gameId: string }) {
    const { data, error } = await client.GET(
      `/api/players/current-player/games/rps`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
        params: {
          query: {
            ids: [rest.gameId],
          },
        },
      },
    );
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    if (data.data?.length) {
      return {
        data: data.data[0],
      };
    }
    return {
      data: null,
    };
  }
  async requestGame({
    token,
    move,
    playerId,
    betAmount,
  }: {
    token: string;
    move: "rock" | "paper" | "scissors";
    playerId: string;
    betAmount?: number;
  }) {
    const { data, error } = await client.POST(`/api/games/rps/requests`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      body: {
        inviting_player_id: playerId,
        move,
        ...(betAmount !== undefined ? { bet_amount: betAmount } : {}),
      },
    });
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  }

  async requestGameEmail({
    token,
    invitingPlayerEmail,
    move,
  }: {
    token: string;
    invitingPlayerEmail: string;
    move: "rock" | "paper" | "scissors";
  }) {
    const { data, error } = await client.POST(
      `/api/games/rps/requests/unregistered`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: {
          "inviting-player-email": invitingPlayerEmail,
          move,
        },
      },
    );
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  }

  async findPlayerByEmail({ token, email }: { token: string; email: string }) {
    const { data, error } = await client.GET(
      `/api/players/registered/email/{inviting-player-email}`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
        params: {
          path: {
            "inviting-player-email": email,
          },
        },
      },
    );
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  }
  async getLedgerBalance({ token }: { token: string }) {
    const { data, error } = await client.GET(`/api/ledger/balance`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data;
  }

  async challengeHouse({
    token,
    move,
    betAmount,
  }: {
    token: string;
    move: "rock" | "paper" | "scissors";
    betAmount?: number;
  }): Promise<ChallengeHouseResult> {
    const { data, error } = await client.POST("/api/games/rps/house", {
      headers: { Authorization: `Bearer ${token}` },
      body: {
        move,
        ...(betAmount !== undefined ? { bet_amount: betAmount } : {}),
      },
    });
    if (error) {
      throw ApiError.fromErrorModel(error);
    }
    return data.data;
  }
}

export const rpsGameQueries = new RpsGameQueries();
