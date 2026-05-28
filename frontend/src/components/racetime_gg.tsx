import { ButtonData } from "./ButtonList"
import { Authorize, GenTokens } from "../../wailsjs/go/main/App";

const DEBUG = true;

const COMPONENT = "RACETIME";

function logRT(message: string, ...args: any[]) {
    console.log(`[INFO] ${COMPONENT}: ${message}`, ...args);
}

function logRTDebug(message: string, ...args: any[]) {
    if (!DEBUG) return;

    console.debug(`[DEBUG] ${COMPONENT}: ${message}`, ...args);
}

function logRTWarn(message: string, ...args: any[]) {
    console.warn(`[WARN] ${COMPONENT}: ${message}`, ...args);
}

function logRTError(message: string, error?: unknown, ...args: any[]) {
    console.error(`[ERROR] ${COMPONENT}: ${message}`, error, ...args);
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
        logRTDebug("fetching race list from %s", restUrl);

        const response = await fetch(restUrl + "/races/data");

        logRTDebug(
            "race list response status=%d ok=%s",
            response.status,
            response.ok,
        );

        if (!response.ok) {
            throw new Error(`unexpected status code ${response.status}`);
        }

        // Read x-date-exact header from response
        const exactHeader = response.headers.get("x-date-exact");

        if (!exactHeader) {
            logRTWarn("missing x-date-exact header");

            throw new Error("missing x-date-exact header");
        }

        const serverTime = new Date(exactHeader);

        logRTDebug("server time=%s", serverTime.toISOString());

        const json = await response.json();

        logRT(
            "received race list count=%d",
            json.races?.length ?? 0,
        );

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

            logRTDebug(
                "race category=%s url=%s entrants=%d finished=%d status=%s startedAt=%s runtime=%s",
                categoryName,
                URL,
                entrantCount,
                entrantFinishedCount,
                status,
                startedAt?.toISOString(),
                runTime,
            );

            DATA.push({
                id: index.toString(),
                URL: URL,
                label:
                    "[" + runTime + "] " +
                    " (" + URL + ") " +
                    categoryName +
                    " - " +
                    goal +
                    " (" +
                    entrantFinishedCount +
                    "/" +
                    entrantCount +
                    " Finished)",
            });
        }

        logRT("race list built count=%d", DATA.length);

        return DATA;
    } catch (err) {
        logRTError("RaceList failed", err);

        return [];
    }
}

// Authenticate and get user tokens
export async function LoginWithOAuth() {
    try {
        logRT("starting oauth flow");

        await Authorize();

        logRT("oauth authorization complete");

        await GenTokens();

        logRT("token generation complete");
    } catch (error) {
        logRTError("OAuth login failed", error);

        throw error;
    }
}