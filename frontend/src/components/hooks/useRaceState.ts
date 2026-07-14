import { useEffect, useMemo, useState } from "react";

import { RaceInfo, UserStatus } from "../../types/racetime";
import { moduleLogger } from "../logger";

const log = moduleLogger("RACE_STATE");

type Props = {
  raceInfo?: RaceInfo;
};

export function useRaceState({ raceInfo }: Props) {
  const [userStatus, setUserStatus] = useState<UserStatus>("not_joined");

  const raceLocked = useMemo(
    () =>
      raceInfo?.Status === "in_progress" ||
      raceInfo?.EndedAt != null ||
      raceInfo?.CancelledAt != null,

    [raceInfo],
  );

  const raceStarted = raceInfo?.Status === "in_progress";

  const joined =
    userStatus === "ready" ||
    userStatus === "not_ready" ||
    userStatus === "in_progress" ||
    userStatus === "done" ||
    userStatus === "dnf" ||
    userStatus === "dq";

  const showJoin = !raceLocked;
  const showReady = joined && !raceLocked;
  const showDone = joined && raceStarted;
  const showForfeit = joined && raceStarted;

  useEffect(() => {
    if (!raceInfo) {
      return;
    }

    const me = raceInfo.Entrants.find(
      (entrant) => entrant.user.id === raceInfo.UserID,
    );

    const nextStatus = me ? (me.status.value as UserStatus) : "not_joined";

    setUserStatus(nextStatus);
  }, [raceInfo]);

  useEffect(() => {
    if (!raceInfo) {
      return;
    }

    if (raceStarted && (userStatus === "ready" || userStatus === "not_ready")) {
      log.info(`race started status=${userStatus}`);

      if (userStatus === "ready") {
        setUserStatus("in_progress");
      } else if (raceInfo.DisqualifyUnready) {
        log.warn("user disqualified for not ready");

        setUserStatus("dq");
      }
    }
  }, [raceInfo, raceStarted, userStatus]);

  return {
    userStatus,
    setUserStatus,

    raceLocked,
    raceStarted,

    joined,

    showJoin,
    showReady,
    showDone,
    showForfeit,
  };
}
