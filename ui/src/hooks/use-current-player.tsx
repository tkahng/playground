import { PlayerContext } from "@/context/player-context";
import { useContext } from "react";

export const usePlayer = () => useContext(PlayerContext);
