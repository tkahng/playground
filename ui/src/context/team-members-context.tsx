// react context for TeamMemberState
import { useLocalStorage } from "@/hooks/use-local-storage";
import { TeamMember } from "@/schema.types";
import React, { createContext, useContext } from "react";

type TeamMemberContextType = {
  currentMember: TeamMember | null;
  setMember: (member: TeamMember | null) => void;
};

const TeamMemberContext = createContext<TeamMemberContextType>({
  currentMember: null,
  setMember: () => {
    throw new Error("setMember function is not implemented");
  },
});

export const TeamMemberProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const [member, setMember] = useLocalStorage<TeamMember | null>(
    "currentMember",
    null
  );

  const values = React.useMemo(() => {
    return {
      currentMember: member,
      setMember,
    };
  }, [member, setMember]);

  return (
    <TeamMemberContext.Provider value={values}>
      {children}
    </TeamMemberContext.Provider>
  );
};

export const useTeamMemberContext = () => useContext(TeamMemberContext);
