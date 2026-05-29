import { ButtonData } from "./ButtonList";
import { moduleLogger } from "./logger";
import { Authorize, GenTokens } from "../../wailsjs/go/main/App";

const log = moduleLogger("RACETIME");

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

        log.debug(
            `race list response status=${response.status} ok=${response.ok}`,
        );

        if (!response.ok) {
            throw new Error(
                `unexpected status code ${response.status}`,
            );
        }

        // Read x-date-exact header from response
        const exactHeader = response.headers.get("x-date-exact");

        if (!exactHeader) {
            log.warn("missing x-date-exact header");

            throw new Error("missing x-date-exact header");
        }

        const serverTime = new Date(exactHeader);

        log.debug(
            `server time=${serverTime.toISOString()}`,
        );

        const json = await response.json();

        log.info(
            `received race list count=${json.races?.length ?? 0}`,
        );

        // Populate buttons with races
        const DATA: ButtonData[] = [];

        for (let index = 0; index < json.races.length; index++) {
            const race = json.races[index];

            const categoryName = race.category.name;
            const URL = race.url;
            const entrantCount = race.entrants_count;
            const entrantFinishedCount =
                race.entrants_count_finished;
            const goal = race.goal.name;
            const status = race.status.value;

            // Convert started_at string to Date
            const startedAt = race.started_at
                ? new Date(race.started_at)
                : null;

            let runTime = "00:00:00";

            if (
                status === "in_progress" &&
                startedAt &&
                !isNaN(startedAt.getTime())
            ) {
                const elapsedMs =
                    serverTime.getTime() - startedAt.getTime();

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

            DATA.push({
                id: index.toString(),
                URL,
                label:
                    `[${runTime}] ` +
                    `(${URL}) ` +
                    `${categoryName} - ` +
                    `${goal} ` +
                    `(${entrantFinishedCount}/${entrantCount} Finished)`,
            });
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