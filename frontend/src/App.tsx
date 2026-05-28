import { useEffect, useState, useRef, useLayoutEffect } from 'react';
import logo from './assets/images/logo-universal.png';
import connected from './assets/images/broadcast_icon_connected.png';
import disconnected from './assets/images/broadcast_icon_disconnected.png';
import './App.css';
import * as racetime from "../wailsjs/go/main/App";
import { LoginWithOAuth, RaceList } from './components/racetime_gg';
import { EventsOn, WindowSetSize } from "../wailsjs/runtime";
import ButtonList, { ButtonData } from "./components/ButtonList"

const DEBUG = true;

const APP_COMPONENT = "APP";

function logApp(message: string, ...args: any[]) {
    if (!DEBUG) return;

    console.log(`[INFO] ${APP_COMPONENT}: ${message}`, ...args);
}

function logAppDebug(message: string, ...args: any[]) {
    if (!DEBUG) return;

    console.debug(`[DEBUG] ${APP_COMPONENT}: ${message}`, ...args);
}

function logAppInfo(message: string, ...args: any[]) {
    console.warn(`[INFO] ${APP_COMPONENT}: ${message}`, ...args);
}

function logAppWarn(message: string, ...args: any[]) {
    console.warn(`[WARN] ${APP_COMPONENT}: ${message}`, ...args);
}

function logAppError(message: string, error?: unknown, ...args: any[]) {
    console.error(`[ERROR] ${APP_COMPONENT}: ${message}`, error, ...args);
}

enum ConnectionStatus {
    Disconnected = 0,
    Connected = 1,
    Reconnecting = 2,
    WaitingForGame,
}

type ConnectionState = {
    connection_status: ConnectionStatus;
    message: string;
};

type UserInfo = {
    ID: string
    FullName: string
    Name: string
    Discriminator: string
    Avatar: string
    TwitchName: string
    IsStaff: boolean
}

type RaceInfo = {
    Version: number
    Goal: string
    Game: string
    RaceID: string
    Info: string
    DisplayResults: boolean
    EntrantCount: number
    EntrantFinishedCount: number
    EntrantInactiveCount: number
    Entrants: Entrant[]
    Text: ChatMessage[]
    Ranked: boolean
    AutoStart: boolean
    StatusVerbose: string
    StatusHelpText: string
    DisqualifyUnready: boolean
}

type ChatMessage = {
    id: string                // "id": "<string>",
    user: User                // "user": { <user info object> },
    bot: string               // "bot": "<string>",
    direct_to: User           // "direct_to": { <user info object> },
    posted_at: string         // "posted_at": "<iso date string>"
    message: string           // "message": "<string>",
    message_plain: string     // "message_plain": "<string>",
    highlight: boolean        // "highlight": <bool>,
    is_dm: boolean            // "is_dm": <bool>,
    is_bot: boolean           // "is_bot": <bool>,
    is_system: boolean        // "is_system": <bool>,
    is_pinned: boolean        // "is_pinned": <bool>,
    delay: string             // "delay": "<iso duration string>",
    //	    "actions" { <action objects> }
}

type User = {
    id: string              // "id": "fR42gLweew3pQlm4",
    full_name: string       // "full_name": "Mario#5527",
    name: string            // "name": "Mario",
    discriminator: string   // "discriminator": "5527",
    url: string             // "url": "/user/fR42gLweew3pQlm4",
    avatar: string          // "avatar": "/media/mario.png",
    pronouns: string        // "pronouns": "he/him",
    flair: string           // "flair": "monitor supporter",
    twitch_name: string     // "twitch_name": "ItsaMeMario",
    twitch_channel: string  // "twitch_channel": "https://www.twitch.tv/itsamemario",
    can_moderate: boolean   // "can_moderate": false
}

type Entrant = {
    user: User                  // user: User data blob for this entrant.
    value: UserStatus           // value: A machine-parsable status text.
    verbose_value: string       // verbose_value: A user-parsable status text, e.g. "In progress".
    help_text: string           // help_text: Describes the status, e.g. "Did not finish the race.".
    finish_time: string         // finish_time: The user's final finish time, or null if they've not finished (ISO 8601 duration).
    finished_at: string         // finished_at: The date/time when the user finished, or null if they've not finished (ISO 8601 date).
    place: number               // place: Integer indicating what position the user finished in.
    place_ordinal: string       // place_ordinal: String ordinal version of place, e.g. "3rd".
    score: number               // score: Integer amount of points earned by this entrant on the relevant leaderboard. Note that this is not the entrant's current score (unless the race is in progress), it is the score they had when they entered the race, not after.
    score_change: number        // score_change Integer amount of points gained/lost as a result of this race, or null (not zero!) if race is not recorded.
    comment: string             // comment: A string containing a pithy comeback supplied by the user post-race, or null if they have no comment. If hide_comments is true and the race has not concluded, this field is always null.
    has_comment: boolean        // has_comment: A boolean indicating if the entrant has made a comment. This field is unaffected by the hide_comments setting.
    stream_live: boolean        // stream_live: Boolean indicating if the user's stream is currently live. This is updated in real-time while a race is in progress, but once an entrant has finished, forfeited or been disqualified it will not be updated.
    stream_override: boolean    // stream_override: Boolean indicating if a moderator overrode the streaming requirement for this race entrant,
}

type UserStatus =
    "requested"     // (requested to join)
    | "invited"     // (invited to join)
    | "declined"    // (declined invitation)
    | "partitioned" // (moved to a 1v1 race room, only for 1v1 ladder races)
    | "not_joined"  // default state (set on leave)
    | "ready"       // only when joined before race start
    | "not_ready"   // only when joined before race start
    | "in_progress" // only when joined after race start
    | "done"        // only when joined after race start
    | "dnf"         // only when joined after race start (did not finish, i.e. forfeited)
    | "dq"          // only when joined after race start (disqualified)

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
    const [textEntry, setTextEntry] = useState<string>("");
    const [joinVisible, setJoinVisible] = useState<boolean>(true);
    const [readyVisible, setReadyVisible] = useState<boolean>(true);
    const [doneVisible, setDoneVisible] = useState<boolean>(true);
    const [forfeitVisible, setForfeitVisible] = useState<boolean>(true);
    const [userStatus, setUserStatus] = useState<UserStatus>("not_joined");
    const [raceInfo, setRaceInfo] = useState<RaceInfo>();
    const [userInfo, setUserInfo] = useState<UserInfo>();
    const [entrantList, setEntrantList] = useState<Entrant[]>([]);
    const [canJoin, setCanJoin] = useState<boolean>(true);
    const [activeChatTab, setActiveChatTab] = useState<string>("main");
    const [unreadTabs, setUnreadTabs] = useState<Set<string>>(new Set());
    const [openSplitConnection, setOpenSplitConnection] =
        useState<ConnectionState>({
            connection_status: ConnectionStatus.Disconnected,
            message: "Opensplit Not Found",
        });

    const joined = userStatus !== "not_joined"

    const raceStarted = disableStatuses.has(userStatus)

    const chatRef = useRef<HTMLDivElement | null>(null);

    const wasAtBottomRef = useRef(true);

    const directMessageUsers = Array.from(
        new Map(
            (raceInfo?.Text ?? [])
                .filter((msg) => msg.is_dm && msg.user)
                .map((msg) => [msg.user.id, msg.user])
        ).values()
    );

    const chatTabs = [
        { id: "main", label: "Main Chat" },
        ...directMessageUsers.map((user) => ({
            id: user.id,
            label: user.name,
        })),
    ];

    const filteredMessages = (raceInfo?.Text ?? []).filter((message) => {
        if (activeChatTab === "main") {
            return !message.is_dm;
        }

        return (
            message.is_dm &&
            message.user?.id === activeChatTab
        );
    });

    useEffect(() => {
        const messages = raceInfo?.Text ?? [];

        const nextUnread = new Set(unreadTabs);

        for (const message of messages) {
            if (
                message.is_dm &&
                message.user &&
                message.user.id !== activeChatTab
            ) {
                nextUnread.add(message.user.id);
            }
        }

        setUnreadTabs(nextUnread);
    }, [raceInfo?.Text, activeChatTab]);


    const isAtBottom = () => {
        const el = chatRef.current;
        if (!el) return true;

        return el.scrollHeight - el.scrollTop - el.clientHeight < 50;
    };

    // track scroll position
    useEffect(() => {
        const el = chatRef.current;
        if (!el) return;

        const onScroll = () => {
            wasAtBottomRef.current =
                el.scrollHeight - el.scrollTop - el.clientHeight < 50;
        };

        el.addEventListener("scroll", onScroll);
        return () => el.removeEventListener("scroll", onScroll);
    }, []);

    // auto-scroll only if user was already at bottom
    useLayoutEffect(() => {
        const el = chatRef.current;
        if (!el) return;

        if (wasAtBottomRef.current) {
            el.scrollTop = el.scrollHeight;
        }
    }, [raceInfo?.Text]);

    const showJoin = !raceStarted
    const showReady = joined && !raceStarted
    const showDone = joined && raceStarted
    const showForfeit = joined && raceStarted

    const hasTwitchName =
        userInfo?.TwitchName != null &&
        userInfo.TwitchName.trim() !== "";

    const handleJoinClick = async () => {
        await racetime.Join(joinVisible)
        logApp("join toggled current=%s", joinVisible);

        if (joinVisible) {
            // joining
            setUserStatus("not_ready")
        } else {
            // leaving
            setUserStatus("not_joined")
        }

        setJoinVisible(!joinVisible)
    }

    const handleReadyClick = async () => {
        await racetime.Ready(readyVisible)
        logApp("ready toggled current=%s", readyVisible);

        if (readyVisible) {
            // becoming ready
            setUserStatus("ready")
        } else {
            // becoming unready
            setUserStatus("not_ready")
        }

        setReadyVisible(!readyVisible)
    }

    const handleDoneClick = async () => {
        await racetime.Done(doneVisible)
        logApp("done toggled current=%s", doneVisible);

        if (doneVisible) {
            setUserStatus("done")
        } else {
            setUserStatus("in_progress")
        }

        setDoneVisible(!doneVisible)
    }

    const handleForfeitClick = async () => {
        await racetime.Forfeit(forfeitVisible)
        logApp("forfeit toggled current=%s", forfeitVisible);

        if (forfeitVisible) {
            setUserStatus("dnf")
        } else {
            setUserStatus("in_progress")
        }

        setForfeitVisible(!forfeitVisible)
    }

    const handleSend = async () => {
        if (!textEntry.trim()) {
            logAppWarn("attempted to send empty message");
            return;
        }

        const id = crypto.randomUUID();

        try {
            logAppDebug(
                "sending chat message id=%s length=%d",
                id,
                textEntry.length,
            );

            await racetime.SendText(textEntry, id);

            logApp("chat message sent successfully");

            setTextEntry("");
        } catch (err) {
            logAppError("SendText failed", err);
        }
    };

    const handleChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const value = event.target.checked;

        await racetime.HideResults(value);
    };

    const getStatusColor = (state: ConnectionStatus) => {
        switch (state) {
            case ConnectionStatus.Disconnected:
                return "red";
            case ConnectionStatus.Connected:
                return "#00FF00";
            case ConnectionStatus.Reconnecting:
                return "yellow";
            case ConnectionStatus.WaitingForGame:
                return "orange";
        }
    };

    useEffect(() => {
        return EventsOn("opensplit:connection", (s: ConnectionState) => {
            setOpenSplitConnection(s);
            logApp(
                "opensplit status=%d message=%s",
                s.connection_status,
                s.message,
            );
        });
    }, []);

    // Sets user to in_progress or dq when race starts
    useEffect(() => {
        if (!raceInfo) return

        const raceStarted =
            raceInfo.StatusVerbose?.toLowerCase().includes("progress") ||
            raceInfo.StatusVerbose?.toLowerCase().includes("started")

        if (
            raceStarted &&
            (userStatus === "ready" || userStatus === "not_ready")
        ) {
            if (userStatus === "ready") {
                setUserStatus("in_progress")
            } else {
                if (raceInfo.DisqualifyUnready) {
                    setUserStatus("dq")
                }
            }
        }
    }, [raceInfo?.StatusVerbose])

    // User can join race
    useEffect(() => {
        const eligibilityEvent = EventsOn("joinEligibility", (eligible: boolean) => {
            setCanJoin(eligible)
        })

        return () => {
            eligibilityEvent()
        }
    }, [])

    // Chat updated
    useEffect(() => {
        const newChatText = EventsOn("chatUpdated", (chatText: ChatMessage[]) => {
            logAppDebug("chat updated messages=%d", chatText.length);
            const shouldAutoScroll = isAtBottom();

            setRaceInfo((prev) => {
                if (!prev) return prev;
                return { ...prev, Text: chatText };
            });

            wasAtBottomRef.current = shouldAutoScroll;
        });

        return () => {
            newChatText();
        };
    }, []);

    // UserInfo updated
    useEffect(() => {
        const newUserInfo = EventsOn("userInfo", (userInfo: UserInfo) => {
            setUserInfo(userInfo)
            logApp(
                "user info updated user=%s twitchLinked=%s",
                userInfo.Name,
                userInfo.TwitchName !== "",
            );
        })
        return () => {
            newUserInfo();
        };
    }, []);

    // RaceInfo updated
    useEffect(() => {
        const newRaceState = EventsOn("raceStateUpdated", (currentRace: RaceInfo) => {
            setRaceInfo(currentRace)
            logApp(
                "race updated goal=%s entrants=%d",
                currentRace.Goal,
                currentRace.EntrantCount,
            );
        })
        return () => {
            newRaceState();
        };
    }, []);

    // Entrants updated
    useEffect(() => {
        const newEntrants = EventsOn("entrantsUpdated", (entrantList: Entrant[]) => {
            setEntrantList(entrantList)
            logAppDebug("entrants updated count=%d", entrantList.length);
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
        logAppInfo("checking stored auth token");
        (
            async () => {
                const raceToken = await racetime.CheckTokens()
                setToken(raceToken)
                logApp("token check complete present=%s", raceToken !== "");
            }
        )()
    }, [])

    // Gets list of races
    useEffect(() => {
        if (token == "") {
            logAppWarn("race polling skipped: token missing");
            return
        }

        if (race != "") {
            logAppWarn("race polling stopped: race joined");
            return
        }

        const fetchRaces = async () => {
            logAppDebug("fetching race list");
            //local dev
            // const raceObj = await RaceList("http://localhost:8000")
            //live
            const raceObj = await RaceList("https://racetime.gg")
            setRaceList(raceObj ?? [])
            logApp("race list updated count=%d", raceObj?.length ?? 0);
        }

        fetchRaces()

        const intervalId = setInterval(() => {
            fetchRaces()
        }, 5000)

        return () => clearInterval(intervalId)
    }, [token, race])

    useEffect(() => {
        logAppDebug("setting window size");

        WindowSetSize(320, 580);
    }, []);

    if (token == "") {
        // no token
        // show login button
        return (
            <div id="Auth">
                <div
                    style={{
                        display: "flex",
                        width: "100%",
                        justifyContent: "center",
                        marginTop: "20px",
                    }}
                    className="status">

                    <table>
                        <tbody>
                            <tr>
                                <td>
                                    <div
                                        style={{
                                            backgroundColor: getStatusColor(
                                                openSplitConnection.connection_status,
                                            ),
                                            borderRadius: "20px",
                                            height: "15px",
                                            width: "15px",
                                        }}
                                    ></div>
                                </td>
                                <td>{openSplitConnection.message}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <button
                    onClick={async () => {
                        await LoginWithOAuth()
                        const raceToken = await racetime.CheckTokens()
                        setToken(raceToken)
                    }}>
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
                    <div
                        style={{
                            display: "flex",
                            width: "100%",
                            justifyContent: "center",
                            marginTop: "20px",
                        }}
                        className="status">

                        <table>
                            <tbody>
                                <tr>
                                    <td>
                                        <div
                                            style={{
                                                backgroundColor: getStatusColor(
                                                    openSplitConnection.connection_status,
                                                ),
                                                borderRadius: "20px",
                                                height: "15px",
                                                width: "15px",
                                            }}
                                        ></div>
                                    </td>
                                    <td>{openSplitConnection.message}</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                    <ButtonList
                        data={raceList}
                        onClick={async (item) => {
                            try {
                                logApp("joining websocket race=%s", item.URL);

                                setJoinedRace(item.URL);

                                await racetime.WebSocketConnection(item.URL);

                                logApp("websocket connected race=%s", item.URL);
                            } catch (err) {
                                logAppError("failed to connect websocket", err);
                            }
                        }}
                    />
                </div>
            )
        } else {
            // race selected
            // show race window
            return (
                <div id="RaceWindow">
                    <div
                        style={{
                            display: "flex",
                            width: "100%",
                            justifyContent: "center",
                            marginTop: "20px",
                        }}
                        className="status">

                        <table>
                            <tbody>
                                <tr>
                                    <td>
                                        <div
                                            style={{
                                                backgroundColor: getStatusColor(
                                                    openSplitConnection.connection_status,
                                                ),
                                                borderRadius: "20px",
                                                height: "15px",
                                                width: "15px",
                                            }}
                                        ></div>
                                    </td>
                                    <td>{openSplitConnection.message}</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                    <button
                        onClick={async () => {
                            logApp("disconnecting from race");
                            await racetime.Join(false)
                            await racetime.DisconnectRace()

                            setJoinVisible(true)
                            setReadyVisible(true)
                            setDoneVisible(true)
                            setForfeitVisible(true)

                            setUserStatus("not_joined")

                            setJoinedRace("")
                            setRaceInfo(undefined)
                            setEntrantList([])
                        }}>
                        Back to Races
                    </button>

                    <div>{"Game: " + raceInfo?.Game}</div>
                    <div>{"Race: " + race}</div>
                    <div>{"Goal: " + raceInfo?.Goal}</div>
                    <div>{"Info: " + raceInfo?.Info}</div>
                    <div>{"Status: " + raceInfo?.StatusVerbose}</div>
                    <div>{raceInfo?.StatusHelpText}</div>

                    <div>{"Ranked: " + (raceInfo?.Ranked ? "Yes" : "No")}</div>

                    <div>{"Auto Start: " + (raceInfo?.AutoStart ? "Enabled" : "Disabled")}</div>
                    <div>
                        {raceInfo?.Entrants?.map((Entrant, index) => (
                            <div key={index}>
                                <img
                                    src={Entrant.stream_live || Entrant.stream_override
                                        ? connected
                                        : disconnected}
                                    alt={Entrant.stream_live || Entrant.stream_override
                                        ? "Connected"
                                        : "Disconnected"}
                                    width={24}
                                    height={24}
                                />

                                <div>{Entrant.place_ordinal}</div>
                                <div>
                                    <img
                                        src={Entrant.user.avatar}
                                        alt={Entrant.user.name}
                                        width={32}
                                        height={32}
                                    />
                                </div>
                                <div>{Entrant.user.name}</div>
                                <div>{Entrant.user.discriminator}</div>
                                <div>{Entrant.user.pronouns}</div>
                                <div>{Entrant.value}</div>
                                <div>{Entrant.finish_time}</div>
                                <div>{Entrant.score_change}</div>
                            </div>
                        ))}
                        <div>{raceInfo?.EntrantCount + " entrants (" + raceInfo?.EntrantInactiveCount + ")"}</div>
                    </div>

                    {/* add scrolling text window */}
                    <div className="chatContainer">

                        {/* Tabs */}
                        <div className="chatTabs">
                            {chatTabs.map((tab) => (
                                <button
                                    key={tab.id}
                                    className={
                                        activeChatTab === tab.id
                                            ? "chatTab active"
                                            : "chatTab"
                                    }
                                    onClick={() => {
                                        setActiveChatTab(tab.id);

                                        setUnreadTabs((prev) => {
                                            const next = new Set(prev);
                                            next.delete(tab.id);
                                            return next;
                                        });
                                    }}>
                                    {tab.label} {unreadTabs.has(tab.id) ? "•" : ""}
                                </button>
                            ))}
                        </div>

                        {/* Chat messages */}
                        <div
                            ref={chatRef}
                            className="chatBox">

                            {filteredMessages.map((message) => (
                                <div
                                    key={message.id}
                                    className={message.is_dm ? "dmMessage" : "mainMessage"}>
                                    <div className="chatMeta">
                                        <span>{message.posted_at}</span>
                                        <span>
                                            {message.user?.name ?? "System"}
                                        </span>
                                    </div>

                                    <div className="chatText">
                                        {message.message}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* add hide results check box */}
                    <label>
                        <input
                            type="checkbox"
                            onChange={handleChange} />
                        Hide Results
                    </label>

                    {/* add force reload button */}
                    {/* <button
                        onClick={async () => { await racetime.ForceReload() }}>
                        Force Reload
                    </button> */}

                    {/* add save log button */}
                    <button
                        onClick={async () => { await racetime.SaveLog() }}>
                        Save Log
                    </button>

                    {/* join button */}
                    <button
                        hidden={!showJoin || !canJoin}
                        disabled={raceStarted || !canJoin || !hasTwitchName}
                        onClick={handleJoinClick}>
                        {joinVisible ? "Join" : "Leave"}
                    </button>

                    {!hasTwitchName && (
                        <div>Please link a Twitch account on racetime.gg to join this race.</div>
                    )}

                    {/* ready button */}
                    <button
                        hidden={!showReady || !canJoin}
                        disabled={!joined || raceStarted}
                        onClick={handleReadyClick}>
                        {readyVisible ? "Ready" : "Unready"}
                    </button>

                    {/* done button */}
                    <button
                        hidden={!showDone}
                        disabled={!joined || !raceStarted}
                        onClick={handleDoneClick}>
                        {!doneVisible ? "Done" : "Undone"}
                    </button>

                    {/* forfeit button */}
                    <button
                        hidden={!showForfeit}
                        disabled={!joined || !raceStarted}
                        onClick={handleForfeitClick}>
                        {!forfeitVisible ? "Forfeit" : "Unforfeit"}
                    </button>

                    {/* add text entry box and send button */}
                    <input
                        value={textEntry}
                        onChange={(e) => setTextEntry(e.target.value)}
                        onKeyDown={(e) => {
                            if (e.key === "Enter") {
                                e.preventDefault();
                                handleSend();
                            }
                        }}
                    />

                    <button onClick={handleSend}>Send</button>
                </div>
            )
        }
    }
}

export default App
