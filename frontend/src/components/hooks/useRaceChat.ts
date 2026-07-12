import { useState } from "react";

import * as racetime from "../../../wailsjs/go/main/App";
import { moduleLogger } from "../logger";

const log = moduleLogger("RACE_CHAT");

export function useRaceChat() {
  const [textEntry, setTextEntry] = useState("");

  const handleSend = async () => {
    if (!textEntry.trim()) {
      log.warn("attempted to send empty chat message");
      return;
    }

    const id = crypto.randomUUID();

    try {
      await racetime.SendText(textEntry, id);

      log.debug(`chat message sent id=${id}`);

      setTextEntry("");
    } catch (err) {
      log.error("SendText failed", err);
    }
  };

  return {
    textEntry,
    setTextEntry,
    handleSend,
  };
}
