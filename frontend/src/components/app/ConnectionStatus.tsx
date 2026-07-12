import {
  ConnectionState,
  ConnectionStatus as ConnectionStatusType,
} from "../../types/racetime";

type Props = {
  connection: ConnectionState;
};

function getStatusColor(state: ConnectionStatusType) {
  switch (state) {
    case ConnectionStatusType.Disconnected:
      return "red";

    case ConnectionStatusType.Connected:
      return "#00FF00";

    case ConnectionStatusType.Reconnecting:
      return "yellow";

    case ConnectionStatusType.WaitingForGame:
      return "orange";
  }
}

export default function ConnectionStatus({ connection }: Props) {
  return (
    <div className="status">
      <table>
        <tbody>
          <tr>
            <td>
              <div
                style={{
                  backgroundColor: getStatusColor(connection.connection_status),
                  borderRadius: "20px",
                  height: "15px",
                  width: "15px",
                }}
              />
            </td>

            <td>{connection.message}</td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}
