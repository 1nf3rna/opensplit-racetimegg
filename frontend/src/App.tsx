import { useEffect, useState } from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import * as racetime from "../wailsjs/go/main/App";
import { LoginWithOAuth, RaceList } from './components/racetime_gg';
import { WindowSetSize } from "../wailsjs/runtime";
import ButtonList, { ButtonData } from "./components/ButtonList"

type UserStatus =
    "ready"
    | "not_ready"
    | "in_progress"
    | "done"
    | "dnf" //(did not finish, i.e. forfeited)
    | "dq" //(disqualified)

const disableStatuses = new Set<UserStatus>([
    "in_progress",
    "done",
    "dnf",
    "dq"
]);

function App() {
    const [token, setToken] = useState<string>("")
    const [raceList, setRaceList] = useState<ButtonData[]>([])
    const [race, setJoinedRace] = useState<string>("")
    const [checked, setChecked] = useState<boolean>(false);
    const [textEntry, setTextEntry] = useState<string>("");
    const [joinVisible, setJoinVisible] = useState<boolean>(true);
    const [readyVisible, setReadyVisible] = useState<boolean>(true);
    const [doneVisible, setDoneVisible] = useState<boolean>(true);
    const [forfeitVisible, setForfeitVisible] = useState<boolean>(true);
    const [userStatus, setUserStatus] = useState<UserStatus>("not_ready");
    const [text, setText] = useState<string>("");
    const [goal, setGoal] = useState<string>("");
    const [raceInfo, setRaceInfo] = useState<string>("Hello from React");
    const [game, setGame] = useState<string>("Hello from React");
    const [entractList, setEntrantList] = useState([]);

    const handleAuthClick =
        async () => {
            await LoginWithOAuth()
            // This just triggers the useeffects functions
            setToken("get tokens")
        };

    const handleJoinClick =
        async () => {
            await racetime.Join(joinVisible)
            setJoinVisible(!joinVisible)
        };

    const handleReadyClick =
        async () => {
            await racetime.Ready(readyVisible)
            setReadyVisible(!readyVisible)
        };

    const handleDoneClick =
        async () => {
            await racetime.Done(doneVisible)
            setDoneVisible(!doneVisible)
        };

    const handleForfeitClick =
        async () => {
            await racetime.Forfeit(forfeitVisible)
            setForfeitVisible(!forfeitVisible)
        };

    const sendToBackend = async () => {
        await racetime.SendText(textEntry);
    };

    const handleChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
        setChecked(checked)
        console.log(event.target.checked);
        await racetime.UpdateEntrantList(checked)
    };

    // Gets tokens from backend
    useEffect(() => {
        // call backend function to get token
        (
            async () => {
                const raceToken = await racetime.CheckTokens()
                setToken(raceToken)
            }
        )()
    }, [])

    // Gets list of races
    useEffect(() => {
        if (token == "") {
            console.log("token is blank\n")
            return
        }

        if (race != "") {
            console.log("race button clicked\n")
            return
        }

        const fetchRaces = async () => {
            const raceObj = await RaceList()
            setRaceList(raceObj ?? [])
        }

        fetchRaces()

        const intervalId = setInterval(() => {
            fetchRaces()
        }, 5000)

        return () => clearInterval(intervalId)
    }, [token, race])

    WindowSetSize(320, 580)
    if (token == "") {
        // no token
        // show login button
        return (
            <div id="Auth">
                <button
                    onClick={() => handleAuthClick}>
                    Racetime.gg Auth
                </button>
            </div>
        )
    } else {
        if (race == "") {
            // no race
            // show race list buttons
            return (
                <div id="RaceList">
                    <ButtonList
                        data={raceList}
                        onClick={(item) => {
                            console.log("Clicked", item);
                            setJoinedRace(item.URL)
                            racetime.WebSocketConnection(item.dataURL)
                        }}
                    />
                </div>
            )
        } else {
            // race selected
            // show race window
            return (
                <div id="RaceWindow">
                    <h1>{"Game: " + game}</h1>
                    <h1>{"Race: " + race}</h1>
                    <h1>{"Goal: " + goal}</h1>
                    <h1>{"Info: " + raceInfo}</h1>
                    <h1>{entractList}</h1>

                    {/* add scrolling text window */}
                    <div
                        style={{
                            width: "400px",
                            height: "150px",
                            overflowY: "auto",
                            border: "1px solid #ccc",
                            padding: "8px",
                            whiteSpace: "pre-wrap",
                        }}>
                        {text}
                    </div>

                    {/* add hide results check box */}
                    <label>
                        <input
                            type="checkbox"
                            checked={checked}
                            onChange={handleChange} />
                        Hide Results
                    </label>

                    {/* add force reload button */}
                    <button
                        onClick={async () => { await racetime.ForceReload() }}>
                        Force Reload
                    </button>

                    {/* add save log button */}
                    <button
                        onClick={async () => { await racetime.SaveLog() }}>
                        Save Log
                    </button>

                    {/* add enter race button */}
                    <button
                        // disable once race starts
                        disabled={disableStatuses.has(userStatus)}
                        onClick={() => handleJoinClick}>
                        {joinVisible ? "Join" : "Leave"}
                    </button>

                    {/* add ready button */}
                    <button
                        // enable once joined; disable on leave
                        disabled={disableStatuses.has(userStatus) || joinVisible}
                        onClick={() => handleReadyClick}>
                        {!readyVisible ? "Ready" : "Unready"}
                    </button>

                    {/* add done button */}
                    <button
                        // only show once race starts
                        disabled={!disableStatuses.has(userStatus)}
                        onClick={() => handleDoneClick}>
                        {!doneVisible ? "Done" : "Undone"}
                    </button>

                    {/* add forfeit button */}
                    <button
                        // only show once race starts
                        disabled={!disableStatuses.has(userStatus)}
                        onClick={() => handleForfeitClick}>
                        {!forfeitVisible ? "Forfeit" : "Unforfeit"}
                    </button>

                    {/* add text entry box and send button */}
                    <input
                        value={textEntry}
                        onChange={(e) => setTextEntry(e.target.value)}
                    />
                    <button onClick={sendToBackend}>Send</button>
                </div>
            )
        }
    }
}

export default App
