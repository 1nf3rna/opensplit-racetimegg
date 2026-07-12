import { ConnectionState } from "../../types/racetime";
import ButtonList, { ButtonData } from "../ButtonList";
import ConnectionStatus from "./ConnectionStatus";

type Props = {
  races: ButtonData[];
  connection: ConnectionState;
  onSelectRace: (url: string) => Promise<void>;
};

export default function RaceSelector({
  races,
  connection,
  onSelectRace,
}: Props) {
  return (
    <div
      id="RaceList"
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100vh",
      }}
    >
      <div
        style={{
          flexShrink: 0,
          display: "flex",
          justifyContent: "center",
          padding: "20px 0 10px",
        }}
      >
        <ConnectionStatus connection={connection} />
      </div>

      <div
        style={{
          flex: 1,
          overflowY: "auto",
          minHeight: 0,
        }}
      >
        <ButtonList data={races} onClick={(item) => onSelectRace(item.URL)} />
      </div>
    </div>
  );
}
