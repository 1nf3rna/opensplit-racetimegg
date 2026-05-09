import { useEffect, useState } from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import * as racetime from "../wailsjs/go/main/App";
import { LoginWithOAuth, RaceList } from './components/racetime_gg';
import { EventsOn, WindowSetSize } from "../wailsjs/runtime";
import ButtonList, { ButtonData } from "./components/ButtonList"

type RaceInfo = {
    Version: number
    Goal: string
    Game: string
    RaceID: string
    Info: string
    DisplayResults: boolean
    EntrantCount: number
    EntrantFinishedCount: number
    Entrants: Entrant[]
    Text: ChatMessage[]
}

type ChatMessage = {
    ID: string                //	    "id": "<string>",
    User: User                //	    "user": { <user info object> },
    Bot: string               //	    "bot": "<string>",
    DirectTo: User           //	    "direct_to": { <user info object> },
    PostedAt: string         //	    "posted_at": "<iso date string>"
    Message: string           //	    "message": "<string>",
    Message_plain: string     //	    "message_plain": "<string>",
    Highlight: boolean        //	    "highlight": <bool>,
    Is_dm: boolean            //	    "is_dm": <bool>,
    Is_bot: boolean           //	    "is_bot": <bool>,
    Is_system: boolean        //	    "is_system": <bool>,
    Is_pinned: boolean        //	    "is_pinned": <bool>,
    Delay: string          //	    "delay": "<iso duration string>",
    //	    "actions" { <action objects> }
}

type User = {
    Id: string              // "id": "fR42gLweew3pQlm4",
    Full_name: string       // "full_name": "Mario#5527",
    Name: string            // "name": "Mario",
    Discriminator: string   // "discriminator": "5527",
    Url: string             // "url": "/user/fR42gLweew3pQlm4",
    Avatar: string          // "avatar": "/media/mario.png",
    Pronouns: string        // "pronouns": "he/him",
    Flair: string           // "flair": "monitor supporter",
    Twitch_name: string     // "twitch_name": "ItsaMeMario",
    Twitch_channel: string  // "twitch_channel": "https://www.twitch.tv/itsamemario",
    Can_moderate: boolean  // "can_moderate": false
}

type Entrant = {
    User: User                            // user: User data blob for this entrant.
    UserStatus: UserStatus              // value: A machine-parsable status text.
    VerboseValue: string          // verbose_value: A user-parsable status text, e.g. "In progress".
    HelpText: string              // help_text: Describes the status, e.g. "Did not finish the race.".
    FinishTime: string    // finish_time: The user's final finish time, or null if they've not finished (ISO 8601 duration).
    FinishedAt: string        // finished_at: The date/time when the user finished, or null if they've not finished (ISO 8601 date).
    Place: number                           // place: Integer indicating what position the user finished in.
    PlaceOrdinal: string                // place_ordinal: String ordinal version of place, e.g. "3rd".
    Score: number                           // score: Integer amount of points earned by this entrant on the relevant leaderboard. Note that this is not the entrant's current score (unless the race is in progress), it is the score they had when they entered the race, not after.
    ScoreChange: number                    // score_change Integer amount of points gained/lost as a result of this race, or null (not zero!) if race is not recorded.
    Comment: string                      // comment: A string containing a pithy comeback supplied by the user post-race, or null if they have no comment. If hide_comments is true and the race has not concluded, this field is always null.
    HasComment: boolean                    // has_comment: A boolean indicating if the entrant has made a comment. This field is unaffected by the hide_comments setting.
    StreamLive: boolean                     // stream_live: Boolean indicating if the user's stream is currently live. This is updated in real-time while a race is in progress, but once an entrant has finished, forfeited or been disqualified it will not be updated.
    StreamOverride: boolean                 // stream_override: Boolean indicating if a moderator overrode the streaming requirement for this race entrant,
}

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
    const [text, setText] = useState<ChatMessage[]>([]);
    // const [goal, setGoal] = useState<string>("");
    const [raceInfo, setRaceInfo] = useState<RaceInfo>();
    // const [game, setGame] = useState<string>("Hello from React");
    const [entrantList, setEntrantList] = useState<Entrant[]>([]);

    // const handleAuthClick =
    //     async () => {
    //         await LoginWithOAuth()
    //         // This just triggers the useeffects functions
    //         setToken("get tokens")
    //     };

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
        const id: string = crypto.randomUUID();
        console.log(id);
        await racetime.SendText(textEntry, id);
    };

    const handleChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
        setChecked(checked)
        console.log(event.target.checked);
        await racetime.HideResults(checked)
    };

    // Chat updated
    useEffect(() => {
        const newChatText = EventsOn("chatUpdated", (chatText: ChatMessage[]) => {
            setText(chatText)
        })
        return () => {
            newChatText();
        };
    }, []);

    // RaceInfo updated
    useEffect(() => {
        const newRaceState = EventsOn("raceStateUpdated", (currentRace: RaceInfo) => {
            setRaceInfo(currentRace)
        })
        return () => {
            newRaceState();
        };
    }, []);

    // Entrants updated
    useEffect(() => {
        const newEntrants = EventsOn("entrantsUpdated", (entrantList: Entrant[]) => {
            setEntrantList(entrantList)
        })
        return () => {
            newEntrants();
        };
    }, []);

    useEffect(() => {
        setRaceInfo((prev) => {
            if (!prev) return prev

            return {
                ...prev,
                Entrants: entrantList,
            }
        })
    }, [entrantList])

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
                    onClick={async () => {
                        await LoginWithOAuth()
                        // This just triggers the useeffects functions
                        setToken("get tokens")
                    }}>
                    {/* onClick={() => handleAuthClick}> */}
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
                    <h1>{"Game: " + raceInfo?.Game}</h1>
                    <h1>{"Race: " + race}</h1>
                    <h1>{"Goal: " + raceInfo?.Goal}</h1>
                    <h1>{"Info: " + raceInfo?.Info}</h1>
                    <div>
                        {raceInfo?.Entrants.map((Entrant, index) => (
                            <div key={index}>{Entrant.User.Name}</div>
                        ))}
                    </div>

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
                        {text.map((message, index) => (
                            <div key={index}>
                                <div>{message.PostedAt}</div>
                                <div>{message.User.Name}</div>
                                <div>{message.Message}</div>
                            </div>
                        ))}
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
