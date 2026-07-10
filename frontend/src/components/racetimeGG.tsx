import { Authorize, GenTokens } from "../../wailsjs/go/main/App";
import { ButtonData } from "./ButtonList";
import { moduleLogger } from "./logger";

const log = moduleLogger("RACETIME");

export function formatDuration(duration?: string | null): string {
  if (!duration) return "";

  const match = duration.match(
    /^P(?:(\d+)D)?T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?$/,
  );

  if (!match) {
    return duration;
  }

  const days = parseInt(match[1] ?? "0", 10);
  const hours = parseInt(match[2] ?? "0", 10);
  const minutes = parseInt(match[3] ?? "0", 10);
  const seconds = Math.floor(parseFloat(match[4] ?? "0"));

  const pad = (n: number) => n.toString().padStart(2, "0");

  if (days > 0) {
    return `${days}d ${hours}:${pad(minutes)}:${pad(seconds)}`;
  }

  if (hours > 0) {
    return `${hours}:${pad(minutes)}:${pad(seconds)}`;
  }

  return `${minutes}:${pad(seconds)}`;
}

function formatElapsed(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);

  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  return [
    hours.toString().padStart(2, "0"),
    minutes.toString().padStart(2, "0"),
    seconds.toString().padStart(2, "0"),
  ].join(":");
}

// Get list of races to be displayed
export async function RaceList(restUrl: string) {
  try {
    log.debug(`fetching race list from ${restUrl}`);

    const response = await fetch(`${restUrl}/races/data`);

    log.debug(`race list response status=${response.status} ok=${response.ok}`);

    if (!response.ok) {
      throw new Error(`unexpected status code ${response.status}`);
    }

    // Read x-date-exact header from response
    const exactHeader = response.headers.get("x-date-exact");

    if (!exactHeader) {
      log.warn("missing x-date-exact header");

      throw new Error("missing x-date-exact header");
    }

    const serverTime = new Date(exactHeader);

    log.debug(`server time=${serverTime.toISOString()}`);

    const json = await response.json();

    log.info(`received race list count=${json.races?.length ?? 0}`);

    // Populate buttons with races
    const DATA: ButtonData[] = [];

    for (let index = 0; index < json.races.length; index++) {
      const race = json.races[index];

      const categoryName = race.category.name;
      const URL = race.url;
      const entrantCount = race.entrants_count;
      const entrantFinishedCount = race.entrants_count_finished;
      const goal = race.goal.name;
      const status = race.status.value;

      // Convert started_at string to Date
      const startedAt = race.started_at ? new Date(race.started_at) : null;

      let runTime = "00:00:00";

      if (
        status === "in_progress" &&
        startedAt &&
        !isNaN(startedAt.getTime())
      ) {
        const elapsedMs = serverTime.getTime() - startedAt.getTime();

        runTime = formatElapsed(elapsedMs);
      }

      log.debug(
        `race category=${categoryName} ` +
          `url=${URL} ` +
          `entrants=${entrantCount} ` +
          `finished=${entrantFinishedCount} ` +
          `status=${status} ` +
          `startedAt=${startedAt?.toISOString()} ` +
          `runtime=${runTime}`,
      );

      const runtimeSeconds =
        status === "in_progress" && startedAt
          ? Math.max(
              0,
              Math.floor((serverTime.getTime() - startedAt.getTime()) / 1000),
            )
          : 0;

      DATA.push({
        id: index.toString(),
        URL,
        category: categoryName,
        runtimeSeconds,
        label:
          `[${runTime}] ` +
          `(${URL}) ` +
          `${categoryName} - ` +
          `${goal} ` +
          ` (${entrantFinishedCount}/${entrantCount} Finished)`,
      });
    }
    DATA.sort((a, b) => {
      // category
      const cat = (a.category ?? "").localeCompare(b.category ?? "");
      if (cat !== 0) return cat;

      // newest first (lowest runtime first)
      const runtime = (a.runtimeSeconds ?? 0) - (b.runtimeSeconds ?? 0);
      if (runtime !== 0) return runtime;

      // alphabetical url
      return a.URL.localeCompare(b.URL);
    });

    let currentCategory = "";

    for (let i = 0; i < DATA.length; i++) {
      const item = DATA[i];

      if (item.category !== currentCategory) {
        currentCategory = item.category ?? "";

        DATA.splice(i, 0, {
          id: `header-${currentCategory}`,
          label: currentCategory,
          URL: "",
          isHeader: true,
        });

        // Skip over the header we just inserted.
        i++;
      }
    }

    log.info(`race list built count=${DATA.length}`);

    return DATA;
  } catch (err) {
    log.error("RaceList failed", err);

    return [];
  }
}

// Authenticate and get user tokens
export async function LoginWithOAuth() {
  try {
    log.info("starting oauth flow");

    await Authorize();

    log.info("oauth authorization complete");

    await GenTokens();

    log.info("token generation complete");
  } catch (error) {
    log.error("OAuth login failed", error);

    throw error;
  }
}
