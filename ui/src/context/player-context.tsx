import { useNullableLocalStorage } from "@/hooks/use-local-storage";
import { Player } from "@/schema.types";
import { createContext } from "react";

type PlayerContextType = {
  player: Player | null;
  setPlayer: (player: Player | null) => void;
};

export const PlayerContext = createContext<PlayerContextType>({
  player: null,
  setPlayer: () => {
    throw new Error("setPlayer function is not implemented");
  },
});

export const PlayerProvider = ({ children }: { children: React.ReactNode }) => {
  const [player, setPlayer] = useNullableLocalStorage<Player>(
    "currentPlayer",
    null
  );

  const setPlayerFunc = (player: Player | null) => {
    setPlayer(player);
  };
  return (
    <PlayerContext.Provider
      value={{
        player,
        setPlayer: setPlayerFunc,
      }}
    >
      {children}
    </PlayerContext.Provider>
  );
};
