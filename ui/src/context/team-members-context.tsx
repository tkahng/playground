// react context for TeamMemberState
import { useLocalStorage } from "@/hooks/use-local-storage";
import { TeamMember } from "@/schema.types";
import React, { createContext, useContext } from "react";

type TeamMemberContextType = {
  currentMember: TeamMember | null;
  members: TeamMember[];
  setCurrentMember: (member: TeamMember | null) => void;
};

const TeamMemberContext = createContext<TeamMemberContextType | undefined>(
  undefined
);

export const TeamMemberProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const [member, setMember] = useLocalStorage<TeamMember | null>(
    "currentMember",
    null
  );
  const [members, setMembers] = useLocalStorage<TeamMember[]>(
    "teamMembers",
    []
  );
  const values = React.useMemo(() => {
    return {};
  }, []);

  return (
    <TeamMemberContext.Provider value={values}>
      {children}
    </TeamMemberContext.Provider>
  );
};

export const useTeamMemberContext = () => useContext(TeamMemberContext);
