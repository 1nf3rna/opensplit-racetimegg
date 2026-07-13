import { useEffect, useState } from "react";
import { EventsOn } from "../../../wailsjs/runtime";
import {
  ChatMessage,
  Entrant,
  RaceActions,
  RaceInfo,
} from "../../types/racetime";
import { moduleLogger } from "../logger";

const log = moduleLogger("RACE_EVENTS");

export function useRaceEvents() {
  const [raceInfo, setRaceInfo] = useState<RaceInfo>();

  const [raceActions, setRaceActions] = useState<RaceActions>({
    canJoin: false,
    joinReason: "",
    joinAction: "join",

    canLeave: false,
    leaveAction: "leave",

    canReady: false,
    readyReason: "",

    canDone: false,
    doneReason: "",

    canForfeit: false,
    forfeitReason: "",

    streamBlocked: false,
  });

  useEffect(() => {
    log.info("subscribing to race events");

    const removeChat = EventsOn("chatUpdated", (chatText: ChatMessage[]) => {
      log.debug(`chat updated count=${chatText.length}`);

      setRaceInfo((prev) => {
        if (!prev) {
          return prev;
        }

        return {
          ...prev,
          Text: chatText,
        };
      });
    });

    const removeRace = EventsOn("raceStateUpdated", (race: RaceInfo) => {
      log.debug(`race updated goal=${race.Goal}`);

      setRaceInfo(race);
    });

    const removeEntrants = EventsOn(
      "entrantsUpdated",
      (entrants: Entrant[]) => {
        log.debug(`entrants updated count=${entrants.length}`);

        setRaceInfo((prev) => {
          if (!prev) {
            return prev;
          }

          return {
            ...prev,
            Entrants: entrants,
          };
        });
      },
    );

    const removeActions = EventsOn(
      "raceActionsUpdated",
      (actions: RaceActions) => {
        log.debug("race actions updated", actions);

        setRaceActions(actions);
      },
    );

    return () => {
      log.debug("unsubscribing from race events");

      removeChat();
      removeRace();
      removeEntrants();
      removeActions();
    };
  }, []);

  return {
    raceInfo,
    setRaceInfo,

    raceActions,
    setRaceActions,
  };
}
