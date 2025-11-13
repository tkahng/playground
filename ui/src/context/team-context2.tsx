// react context for TeamMemberState
import { useNullableLocalStorage } from "@/hooks/use-local-storage";
import { useTeamBySlugQuery } from "@/lib/queries";
import { TeamMember, TeamWithMember } from "@/schema.types";
import React, { createContext } from "react";

type TeamContextType = {
  team: TeamWithMember | null;
  teamMember: TeamMember | null;
  setTeam: (team: TeamWithMember | null) => void;
};

export const TeamContext2 = createContext<TeamContextType>({
  team: null,
  teamMember: null,
  setTeam: () => {
    throw new Error("setTeam function is not implemented");
  },
});

export const TeamProvider2 = ({ children }: { children: React.ReactNode }) => {
  const { data } = useTeamBySlugQuery();
  const [team, setTeam] = useNullableLocalStorage<TeamWithMember>(
    "currentTeamWithMember",
    data || null
  );
  const [teamMember, setTeamMember] = useNullableLocalStorage<TeamMember>(
    "currentTeamMember",
    data?.member || null
  );
  const setTeamFunc = (team: TeamWithMember | null) => {
    setTeamMember(team?.member || null);
    setTeam(team);
  };
  return (
    <TeamContext2.Provider
      value={{ team, setTeam: setTeamFunc, teamMember: teamMember || null }}
    >
      {children}
    </TeamContext2.Provider>
  );
};
