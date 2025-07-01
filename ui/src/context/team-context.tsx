// react context for TeamMemberState
import { useLocalStorage } from "@/hooks/use-local-storage";
import { Team, TeamMember } from "@/schema.types";
import React, { createContext } from "react";

type TeamContextType = {
  team: Team | null;
  teamMember: TeamMember | null;
  setTeam: (team: Team | null) => void;
};

export const TeamContext = createContext<TeamContextType>({
  team: null,
  teamMember: null,
  setTeam: () => {
    throw new Error("setTeam function is not implemented");
  },
});

export const TeamProvider = ({ children }: { children: React.ReactNode }) => {
  const [team, setTeam] = useLocalStorage<Team | null>("currentTeam", null);
  const [teamMember, setTeamMember] = useLocalStorage<TeamMember | null>(
    "currentTeamMember",
    null
  );
  const setTeamFunc = (team: Team | null) => {
    const member = team?.members?.find((m) => m);
    setTeamMember(member || null);
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
