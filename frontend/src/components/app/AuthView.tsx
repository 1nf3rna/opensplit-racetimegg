import { ConnectionState } from "../../types/racetime";
import ConnectionStatus from "./ConnectionStatus";

type Props = {
  connection: ConnectionState;
  onLogin: () => Promise<void>;
};

export default function AuthView({ connection, onLogin }: Props) {
  return (
    <div id="Auth">
      <div
        style={{
          display: "flex",
          width: "100%",
          justifyContent: "center",
          marginTop: "20px",
        }}
      >
        <ConnectionStatus connection={connection} />
      </div>

      <button onClick={onLogin}>Racetime.gg Auth</button>
    </div>
  );
}
