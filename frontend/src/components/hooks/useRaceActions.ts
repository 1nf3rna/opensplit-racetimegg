import {
  DeclineInvite,
  Done,
  Forfeit,
  Join,
  Leave,
  Ready,
  RequestInvite,
} from "../../../wailsjs/go/app/App";
import { RaceActions, UserStatus } from "../../types/racetime";
import { moduleLogger } from "../logger";

const log = moduleLogger("RACE_ACTIONS");

type Props = {
  actions: RaceActions;

  readyVisible: boolean;
  doneVisible: boolean;
  forfeitVisible: boolean;

  setReadyVisible: (value: boolean) => void;
  setDoneVisible: (value: boolean) => void;
  setForfeitVisible: (value: boolean) => void;

  setUserStatus: (status: UserStatus) => void;
};

export function useRaceActions({
  actions,

  readyVisible,
  doneVisible,
  forfeitVisible,

  setReadyVisible,
  setDoneVisible,
  setForfeitVisible,

  setUserStatus,
}: Props) {
  const handleJoin = async () => {
    log.info("join clicked");

    switch (actions.joinAction) {
      case "join":
      case "accept_invite":
        await Join();
        break;

      case "request_invite":
        await RequestInvite(true);
        break;
    }
  };

  const handleLeave = async () => {
    log.info("leave clicked");

    switch (actions.leaveAction) {
      case "leave":
        await Leave();
        break;

      case "decline_invite":
        await DeclineInvite();
        break;

      case "cancel_invite":
        await RequestInvite(false);
        break;
    }
  };

  const handleReady = async () => {
    log.info(`ready clicked visible=${readyVisible}`);

    await Ready(readyVisible);

    if (readyVisible) {
      setUserStatus("ready");
    } else {
      setUserStatus("not_ready");
    }

    setReadyVisible(!readyVisible);
  };

  const handleDone = async () => {
    log.info(`done clicked visible=${doneVisible}`);

    await Done(doneVisible);

    if (doneVisible) {
      setUserStatus("done");
    } else {
      setUserStatus("in_progress");
    }

    setDoneVisible(!doneVisible);
  };

  const handleForfeit = async () => {
    log.info(`forfeit clicked visible=${forfeitVisible}`);

    await Forfeit(forfeitVisible);

    if (forfeitVisible) {
      setUserStatus("dnf");
    } else {
      setUserStatus("in_progress");
    }

    setForfeitVisible(!forfeitVisible);
  };

  return {
    handleJoin,
    handleLeave,
    handleReady,
    handleDone,
    handleForfeit,
  };
}
