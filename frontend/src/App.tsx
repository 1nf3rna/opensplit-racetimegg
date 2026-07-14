import "./App.css";

import { useState } from "react";

import AuthView from "./components/app/AuthView";
import RaceSelector from "./components/app/RaceSelector";
import RaceView from "./components/app/RaceView";
import { useOpenSplitConnection } from "./components/hooks/useOpenSplitConnection";
import { useRaceActions } from "./components/hooks/useRaceActions";
import { useRaceChat } from "./components/hooks/useRaceChat";
import { useRaceConnection } from "./components/hooks/useRaceConnection";
import { useRaceEvents } from "./components/hooks/useRaceEvents";
import { useRacePolling } from "./components/hooks/useRacePolling";
import { useRaceState } from "./components/hooks/useRaceState";
import { useWindowLayout } from "./components/hooks/useWindowLayout";
import { moduleLogger } from "./components/logger";
import { LoginWithOAuth } from "./components/racetimeGG";

const log = moduleLogger("APP");

function App() {
  const [race, setJoinedRace] = useState<string>("");
  const [hideResults, setHideResults] = useState(false);
  const [readyVisible, setReadyVisible] = useState(true);
  const [doneVisible, setDoneVisible] = useState(true);
  const [forfeitVisible, setForfeitVisible] = useState(true);

  const openSplitConnection = useOpenSplitConnection();

  const {
    raceInfo,
    setRaceInfo,

    raceActions,
  } = useRaceEvents();

  const { token, raceList, refreshToken } = useRacePolling(race);

  const {
    setUserStatus,

    showJoin,
    showReady,
    showDone,
    showForfeit,
  } = useRaceState({
    raceInfo,
  });

  const { handleJoin, handleLeave, handleReady, handleDone, handleForfeit } =
    useRaceActions({
      actions: raceActions,

      readyVisible,
      doneVisible,
      forfeitVisible,

      setReadyVisible,
      setDoneVisible,
      setForfeitVisible,

      setUserStatus,
    });

  const { textEntry, setTextEntry, handleSend } = useRaceChat();

  const resetRaceState = () => {
    setReadyVisible(true);
    setDoneVisible(true);
    setForfeitVisible(true);

    setUserStatus("not_joined");

    setRaceInfo(undefined);
  };

  const { handleSelectRace, handleBack } = useRaceConnection({
    setRace: setJoinedRace,
    resetRaceState,
  });

  const handleLogin = async () => {
    await LoginWithOAuth();
    await refreshToken();
  };

  useWindowLayout(race);

  // no token
  // show login button
  if (!token) {
    return <AuthView connection={openSplitConnection} onLogin={handleLogin} />;
  }

  if (!race) {
    return (
      <RaceSelector
        races={raceList}
        connection={openSplitConnection}
        onSelectRace={handleSelectRace}
      />
    );
  }

  // race selected
  // show race window
  return (
    <RaceView
      race={race}
      raceInfo={raceInfo}
      entrants={raceInfo?.Entrants ?? []}
      connection={openSplitConnection}
      actions={raceActions}
      hideResults={hideResults}
      setHideResults={setHideResults}
      readyVisible={readyVisible}
      doneVisible={doneVisible}
      forfeitVisible={forfeitVisible}
      showJoin={showJoin}
      showReady={showReady}
      showDone={showDone}
      showForfeit={showForfeit}
      onBack={handleBack}
      onJoin={handleJoin}
      onLeave={handleLeave}
      onReady={handleReady}
      onDone={handleDone}
      onForfeit={handleForfeit}
      textEntry={textEntry}
      setTextEntry={setTextEntry}
      onSend={handleSend}
    />
  );
}

export default App;
