import { useContext } from "react";
import { TeamContext } from "../context/team-context";

export const useTeam = () => useContext(TeamContext);
