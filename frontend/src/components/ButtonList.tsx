import { moduleLogger } from "./logger";

const log = moduleLogger("BUTTONS");

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

export default function ButtonList({
  data,
  onClick,
  className,
}: Props) {
  log.debug(`rendering button list count=${data.length}`);

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
            log.info(`button clicked id=${item.id}`);

            if (item.disabled) {
              log.warn(`blocked click for disabled button id=${item.id}`,);

              return;
            }

            try {
              onClick?.(item);

              log.debug(`button handler executed id=${item.id}`,);
            } catch (error) {
              log.error(`button handler failed id=${item.id}`, error,);
            }
          }}
        >
          {item.label}
        </button>
      ))}
    </div>
  );
}