import connected from "../../assets/images/broadcast_icon_connected.png";
import disconnected from "../../assets/images/broadcast_icon_disconnected.png";
import { Entrant } from "../../types/racetime";
import { formatDuration } from "../racetimeGG";

type Props = {
  entrants: Entrant[];
  hideResults: boolean;
};

export default function EntrantList({ entrants, hideResults }: Props) {
  return (
    <div className="entrantList">
      {entrants.map((entrant, index) => (
        <div key={index} className="entrantRow">
          <img
            src={
              entrant.stream_live || entrant.stream_override
                ? connected
                : disconnected
            }
            alt="stream"
            width={16}
            height={16}
          />

          <img src={entrant.user.avatar} width={24} height={24} alt="avatar" />

          {!hideResults && <span>{entrant.place_ordinal}</span>}

          <span>{entrant.user.name}</span>

          {!hideResults && <span>{entrant.status.verbose_value}</span>}

          {!hideResults && <span>{formatDuration(entrant.finish_time)}</span>}
        </div>
      ))}

      <div className="entrantSummary">{entrants.length} entrants</div>
    </div>
  );
}
