import { useEffect, useState } from "react";

import { CheckTokens } from "../../../wailsjs/go/app/App";
import { ButtonData } from "../ButtonList";
import { moduleLogger } from "../logger";
import { RaceList } from "../racetimeGG";

const log = moduleLogger("RACE_POLLING");

export function useRacePolling(activeRace: string) {
  const [token, setToken] = useState("");

  const [raceList, setRaceList] = useState<ButtonData[]>([]);

  useEffect(() => {
    log.info("checking stored auth token");

    (async () => {
      const raceToken = await CheckTokens();

      setToken(raceToken);

      log.info(`token present=${raceToken !== ""}`);
    })();
  }, []);

  useEffect(() => {
    if (!token) {
      log.debug("race polling skipped token missing");

      return;
    }

    if (activeRace) {
      log.debug(`race polling paused activeRace=${activeRace}`);

      return;
    }

    const fetchRaces = async () => {
      log.debug("fetching race list");

      const races = await RaceList("https://racetime.gg");

      setRaceList(races ?? []);

      log.debug(`race list count=${races?.length ?? 0}`);
    };

    fetchRaces();

    const interval = setInterval(fetchRaces, 5000);

    return () => {
      clearInterval(interval);
    };
  }, [token, activeRace]);

  const refreshToken = async () => {
    const raceToken = await CheckTokens();

    setToken(raceToken);
  };

  return {
    token,
    raceList,
    refreshToken,
  };
}
