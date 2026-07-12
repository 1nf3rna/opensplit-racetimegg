import * as racetime from "../../../wailsjs/go/main/App";
import { LogError, LogInfo } from "../../../wailsjs/runtime";

import { moduleLogger } from "../logger";

const log = moduleLogger("RACE_CONNECTION");

type Props = {
  setRace: (race: string) => void;
  resetRaceState: () => void;
};

export function useRaceConnection({ setRace, resetRaceState }: Props) {
  const handleSelectRace = async (url: string) => {
    try {
      LogInfo(`joining websocket race=${url}`);

      setRace(url);

      await racetime.WebSocketConnection(url);

      LogInfo(`websocket connected race=${url}`);
    } catch (err) {
      LogError(`failed to connect websocket: ${err}`);

      setRace("");
    }
  };

  const handleBack = async () => {
    log.info("disconnecting from race");

    await racetime.DisconnectRace();

    setRace("");

    resetRaceState();
  };

  return {
    handleSelectRace,
    handleBack,
  };
}
