import { ButtonData } from "./ButtonList"
import { Authorize, GenTokens } from "../../wailsjs/go/main/App";

const DEBUG = true;

function logRT(message: string, ...args: any[]) {
    if (!DEBUG) return;

    console.log(`[RACETIME] ${message}`, ...args);
}

function logRTError(message: string, error: unknown) {
    console.error(`[RACETIME] ${message}`, error);
}

// Get list of races to be displayed
export async function RaceList(restUrl: string) {
    try {
        logRT("fetching race list from %s", restUrl);
        const response = await fetch(restUrl + "/races/data");
        logRT("race list response status=%d", response.status)
        const json = await response.json();   // parse JSON
        logRT("received %d races", json.races.length);

        // Populate buttons with races
        const DATA: ButtonData[] = [
        ];

        for (let index = 0; index < json.races.length; index++) {
            const categoryName = json.races[index].category.name;
            const URL = json.races[index].url;
            const entrantCount = json.races[index].entrants_count;
            const entrantFinishedCount = json.races[index].entrants_count_finished;
            const goal = json.races[index].goal.name;
            const status = json.races[index].status.value;
            // time stamp format 2025-12-06T08:18:13.788Z
            const startedAt = json.races[index].started_at;
            
            logRT(
                "race category=%s url=%s entrants=%d finished=%d status=%s startedAt=%s",
                categoryName,
                URL,
                entrantCount,
                entrantFinishedCount,
                status,
                startedAt,
            );

            // TODO: this should be saved from the racelist call
            const x_date_exact_header: Date = new Date("2025-12-06T23:01:07Z");
            // var elapsedTime: Date = new Date(x_date_exact_header.getTime() - startedAt.getTime())
            var elapsedTime: Date = new Date(0)
            var runTime = status == 'in_progress' ? elapsedTime : "0"
            DATA.push({
                id: index.toString(),
                URL: URL,
                label: "[" + runTime + "] " + " (" + URL + ") " + categoryName + " - " + goal + " (" + entrantFinishedCount + "/" + entrantCount + " Finished)"
            });
        }

        return DATA
    } catch (err) {
        logRTError("RaceList failed", err);
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
    }
}