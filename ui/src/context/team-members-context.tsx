// react context for TeamMemberState
import { useLocalStorage } from "@/hooks/use-local-storage";
import { TeamMemberState } from "@/schema.types";
import React, { createContext, useContext } from "react";

type TeamMemberContextType = {
  state: TeamMemberState;
  setCurrentMember: (member: TeamMemberState["currentMember"]) => void;
};

const TeamMemberContext = createContext<TeamMemberContextType | undefined>(
  undefined
);

export const TeamMemberProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const [state, setState] = useLocalStorage<TeamMemberState>(
    "teamMemberState",
    {
      currentMember: null,
      members: [],
    }
  );

  const setCurrentMember = (member: TeamMemberState["currentMember"]) => {
    setState({
      ...state,
      currentMember: member,
    });
  };

  return (
    <TeamMemberContext.Provider value={{ state, setCurrentMember }}>
      {children}
    </TeamMemberContext.Provider>
  );
};

export const useTeamMemberContext = () => {
  const context = useContext(TeamMemberContext);
  if (!context) {
    throw new Error(
      "useTeamMemberContext must be used within a TeamMemberProvider"
    );
  }
  return context;
};
