import { ConnectRace, DisconnectRace } from "../../../wailsjs/go/app/App";
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

      try {
        await ConnectRace(url);
      } catch (err) {
        setRace("");
      }
    } catch (err) {
      LogError(`failed to connect websocket: ${err}`);

      setRace("");
    }
  };

  const handleBack = async () => {
    log.info("disconnecting from race");

    await DisconnectRace();

    setRace("");

    resetRaceState();
  };

  return {
    handleSelectRace,
    handleBack,
  };
}
