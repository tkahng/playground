// react context for TeamMemberState
import { Team } from "@/schema.types";
import { createContext } from "react";

type TeamContextType = {
  team: Team | null;
  // setTeam: (team: Team | null) => void;
};

export const TeamContext = createContext<TeamContextType>({
  team: null,
  // setTeam: () => {
  //   throw new Error("setTeam function is not implemented");
  // },
});

// export const TeamProvider = ({ children }: { children: React.ReactNode }) => {
//   const [team, setTeam] = useLocalStorage<Team | null>("currentTeam", null);
//   const values = React.useMemo(() => {
//     return {
//       team,
//       setTeam,
//     };
//   }, [team, setTeam]);

//   return <TeamContext.Provider value={values}>{children}</TeamContext.Provider>;
// };
