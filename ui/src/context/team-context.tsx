// react context for TeamMemberState
import { useNullableLocalStorage } from "@/hooks/use-local-storage";
import { TeamMember, TeamWithMember } from "@/schema.types";
import React, { createContext } from "react";

type TeamContextType = {
  team: TeamWithMember | null;
  teamMember: TeamMember | null;
  setTeam: (team: TeamWithMember | null) => void;
};

export const TeamContext = createContext<TeamContextType>({
  team: null,
  teamMember: null,
  setTeam: () => {
    throw new Error("setTeam function is not implemented");
  },
});

export const TeamProvider = ({ children }: { children: React.ReactNode }) => {
  const [team, setTeam] = useNullableLocalStorage<TeamWithMember>(
    "currentTeamWithMember",
    null
  );
  const [teamMember, setTeamMember] = useNullableLocalStorage<TeamMember>(
    "currentTeamMember",
    null
  );
  const setTeamFunc = (team: TeamWithMember | null) => {
    setTeamMember(team?.member || null);
    setTeam(team);
  };
  return (
    <TeamContext.Provider
      value={{ team, setTeam: setTeamFunc, teamMember: teamMember || null }}
    >
      {children}
    </TeamContext.Provider>
  );
};
