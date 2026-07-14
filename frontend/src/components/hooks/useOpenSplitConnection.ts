import { useEffect, useState } from "react";

import { EventsOn } from "../../../wailsjs/runtime";
import { ConnectionState, ConnectionStatus } from "../../types/racetime";
import { moduleLogger } from "../logger";

const log = moduleLogger("OPENSPLIT_CONNECTION");

const defaultConnection: ConnectionState = {
  connection_status: ConnectionStatus.Disconnected,

  message: "Opensplit Not Found",
};

export function useOpenSplitConnection() {
  const [connection, setConnection] =
    useState<ConnectionState>(defaultConnection);

  useEffect(() => {
    log.info("subscribing to opensplit connection");

    const unsubscribe = EventsOn(
      "opensplit:connection",
      (state: ConnectionState) => {
        log.debug(`opensplit connection status=${state.connection_status}`);

        setConnection(state);
      },
    );

    return () => {
      log.debug("unsubscribing from opensplit connection");

      unsubscribe();
    };
  }, []);

  return connection;
}
