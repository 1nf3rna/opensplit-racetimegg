import { SaveLog } from "../../../wailsjs/go/app/App";
import { RaceActions as RaceActionsType } from "../../types/racetime";

type Props = {
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

  onJoin: () => Promise<void>;
  onLeave: () => Promise<void>;

  onReady: () => Promise<void>;
  onDone: () => Promise<void>;
  onForfeit: () => Promise<void>;
};

export default function RaceActions({
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

  onJoin,
  onLeave,

  onReady,
  onDone,
  onForfeit,
}: Props) {
  const handleSaveLog = async () => {
    await SaveLog();
  };
  return (
    <div className="actionPanel">
      <label>
        <input
          type="checkbox"
          checked={hideResults}
          onChange={(e) => setHideResults(e.target.checked)}
        />
        Hide Results
      </label>

      <button onClick={handleSaveLog}>Save Log</button>

      {actions.canJoin && (
        <button onClick={onJoin}>
          {
            {
              join: "Join",
              accept_invite: "Accept Invite",
              request_invite: "Request Invite",
            }[actions.joinAction]
          }
        </button>
      )}

      {!actions.canJoin && showJoin && actions.joinReason && (
        <div className="hint">{actions.joinReason}</div>
      )}

      {actions.canLeave && (
        <button onClick={onLeave}>
          {
            {
              leave: "Leave",
              decline_invite: "Decline Invite",
              cancel_invite: "Cancel Invite",
            }[actions.leaveAction]
          }
        </button>
      )}

      <button
        disabled={!actions.canReady}
        hidden={!showReady}
        onClick={onReady}
      >
        {readyVisible ? "Ready" : "Unready"}
      </button>

      {!actions.canReady && showReady && actions.readyReason && (
        <div className="hint">{actions.readyReason}</div>
      )}

      <button disabled={!actions.canDone} hidden={!showDone} onClick={onDone}>
        {!doneVisible ? "Done" : "Undone"}
      </button>

      {!actions.canDone && showDone && actions.doneReason && (
        <div className="hint">{actions.doneReason}</div>
      )}

      <button
        disabled={!actions.canForfeit}
        hidden={!showForfeit}
        onClick={onForfeit}
      >
        {!forfeitVisible ? "Forfeit" : "Unforfeit"}
      </button>

      {!actions.canForfeit && showForfeit && actions.forfeitReason && (
        <div className="hint">{actions.forfeitReason}</div>
      )}
    </div>
  );
}
