import {
  ConnectionState,
  Entrant,
  RaceActions as RaceActionsType,
  RaceInfo,
} from "../../types/racetime";
import ChatPanel from "../chat/ChatPanel";
import EntrantList from "../race/EntrantList";
import RaceActions from "../race/RaceActions";
import ConnectionStatus from "./ConnectionStatus";

type Props = {
  race: string;

  raceInfo?: RaceInfo;

  entrants: Entrant[];

  connection: ConnectionState;

  actions: RaceActionsType;

  hideResults: boolean;
  setHideResults: (value: boolean) => void;

  readyVisible: boolean;
  doneVisible: boolean;
  forfeitVisible: boolean;

  showJoin: boolean;
  showReady: boolean;
  showDone: boolean;
  showForfeit: boolean;

  onBack: () => Promise<void>;

  onJoin: () => Promise<void>;
  onLeave: () => Promise<void>;

  onReady: () => Promise<void>;
  onDone: () => Promise<void>;
  onForfeit: () => Promise<void>;

  textEntry: string;
  setTextEntry: (value: string) => void;

  onSend: () => Promise<void>;
};

export default function RaceView({
  race,
  raceInfo,
  entrants,

  connection,

  actions,

  hideResults,
  setHideResults,

  readyVisible,
  doneVisible,
  forfeitVisible,

  showJoin,
  showReady,
  showDone,
  showForfeit,

  onBack,

  onJoin,
  onLeave,

  onReady,
  onDone,
  onForfeit,

  textEntry,
  setTextEntry,
  onSend,
}: Props) {
  return (
    <div id="RaceWindow">
      <div className="raceHeader" />

      <div className="raceMain">
        <div className="raceLeft">
          <div className="raceInfoBlock">
            <div className="raceInfoRow">
              <span className="label">Game:</span>

              <span className="value">{raceInfo?.Game}</span>
            </div>

            <div className="raceInfoRow">
              <span className="label">Race:</span>

              <span className="value">{race}</span>
            </div>

            <div className="raceInfoRow">
              <span className="label">Goal:</span>

              <span className="value">{raceInfo?.Goal}</span>
            </div>

            <div className="raceInfoRow">
              <span className="label">Info:</span>

              <div className="value">{raceInfo?.Info}</div>
            </div>
          </div>

          <ChatPanel messages={raceInfo?.Text ?? []} />
        </div>

        <div className="entrantPanel">
          <div className="raceStatusPanel">
            <button className="backButton" onClick={onBack}>
              Back to Races
            </button>

            <ConnectionStatus connection={connection} />
          </div>

          <div className="raceStatusPanel">
            <div className="timerDisplay">{raceInfo?.StatusVerbose}</div>

            <div className="timerStatus">{raceInfo?.StatusHelpText}</div>

            <div>Ranked: {raceInfo?.Ranked ? "Yes" : "No"}</div>

            <div>
              Auto Start: {raceInfo?.AutoStart ? "Enabled" : "Disabled"}
            </div>
          </div>

          <EntrantList entrants={entrants} hideResults={hideResults} />
        </div>
      </div>

      <RaceActions
        actions={actions}
        hideResults={hideResults}
        setHideResults={setHideResults}
        readyVisible={readyVisible}
        doneVisible={doneVisible}
        forfeitVisible={forfeitVisible}
        showJoin={showJoin}
        showReady={showReady}
        showDone={showDone}
        showForfeit={showForfeit}
        onJoin={onJoin}
        onLeave={onLeave}
        onReady={onReady}
        onDone={onDone}
        onForfeit={onForfeit}
      />

      <div className="chatInputBar">
        <input
          value={textEntry}
          onChange={(e) => setTextEntry(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault();
              onSend();
            }
          }}
        />

        <button onClick={onSend}>Send</button>
      </div>
    </div>
  );
}
