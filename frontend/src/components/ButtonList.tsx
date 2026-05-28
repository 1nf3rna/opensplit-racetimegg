const DEBUG = true;

const COMPONENT = "BUTTONS";

function logButtons(message: string, ...args: any[]) {
  console.log(`[INFO] ${COMPONENT}: ${message}`, ...args);
}

function logButtonsDebug(message: string, ...args: any[]) {
  if (!DEBUG) return;

  console.debug(`[DEBUG] ${COMPONENT}: ${message}`, ...args);
}

function logButtonsWarn(message: string, ...args: any[]) {
  console.warn(`[WARN] ${COMPONENT}: ${message}`, ...args);
}

function logButtonsError(message: string, error?: unknown, ...args: any[]) {
  console.error(`[ERROR] ${COMPONENT}: ${message}`, error, ...args);
}

export type ButtonData = {
  id: string;
  label: string;
  URL: string;
  action?: string;
  disabled?: boolean;
  title?: string;
};

type Props = {
  data: ButtonData[];
  onClick?: (item: ButtonData) => void;
  className?: string;
};

export default function ButtonList({ data, onClick, className }: Props) {
  logButtonsDebug("rendering button list count=%d", data.length);

  return (
    <div className={className} role="group" aria-label="button list">
      {data.map((item) => (
        <button
          key={item.id}
          type="button"
          id={`btn-${item.id}`}
          title={item.title}
          disabled={item.disabled}
          onClick={() => {
            logButtons(
              "button clicked id=%s disabled=%s label=%s",
              item.id,
              item.disabled,
              item.label,
            );

            if (item.disabled) {
              logButtonsWarn(
                "blocked click for disabled button id=%s",
                item.id,
              );

              return;
            }

            try {
              onClick?.(item);

              logButtonsDebug(
                "button handler executed id=%s",
                item.id,
              );
            } catch (error) {
              logButtonsError(
                "button handler failed id=%s",
                error,
                item.id,
              );
            }
          }}
        >
          {item.label}
        </button>
      ))}
    </div>
  );
}