import { useEffect, useState, useRef, useLayoutEffect } from 'react';
import logo from './assets/images/logo-universal.png';
import connected from './assets/images/broadcast_icon_connected.png';
import disconnected from './assets/images/broadcast_icon_disconnected.png';
import './App.css';
import * as racetime from "../wailsjs/go/main/App";
import { LoginWithOAuth, RaceList } from './components/racetime_gg';
import { EventsOn, LogError, LogInfo, WindowSetSize } from "../wailsjs/runtime";
import ButtonList, { ButtonData } from "./components/ButtonList"
import { moduleLogger } from "./components/logger";

const log = moduleLogger("APP");

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
    twitch_name: string
    IsStaff: boolean
}

type RaceInfo = {
    Version: number
    Goal: string
    Game: string
    RaceID: string
    Info: string
    StreamingRequired: boolean
    DisplayResults: boolean
    EntrantCount: number
    EntrantFinishedCount: number
    EntrantInactiveCount: number
    Entrants: Entrant[]
    Text: ChatMessage[]
    Ranked: boolean
    AutoStart: boolean
    Delay: number
    Status: string
    StatusVerbose: string
    StatusHelpText: string
    DisqualifyUnready: boolean
    EndedAt: string | null
    CancelledAt: string | null
}

type ChatMessage = {
    id: string
    user: User
    bot: string
    direct_to: User
    posted_at: string
    message: string
    message_plain: string
    highlight: boolean
    is_dm: boolean
    is_bot: boolean
    is_system: boolean
    is_pinned: boolean
    delay: string
}

type User = {
    id: string
    full_name: string
    name: string
    discriminator: string
    url: string
    avatar: string
    pronouns: string
    flair: string
    twitch_name: string
    twitch_channel: string
    can_moderate: boolean
}

type Entrant = {
    user: User
    value: UserStatus
    verbose_value: string
    help_text: string
    finish_time: string
    finished_at: string
    place: number
    place_ordinal: string
    score: number
    score_change: number
    comment: string
    has_comment: boolean
    stream_live: boolean
    stream_override: boolean
}

type UserStatus =
    "requested"
    | "invited"
    | "declined"
    | "partitioned"
    | "not_joined"
    | "ready"
    | "not_ready"
    | "in_progress"
    | "done"
    | "dnf"
    | "dq"

function App() {
    log.debug("rendering app");

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

    const raceStarted = raceInfo?.Status === "in_progress"
    const joined = userStatus !== "not_joined"

    const raceEndedOrCancelled = raceInfo?.EndedAt != null || raceInfo?.CancelledAt != null

    const raceInProgress = raceInfo?.Status === "in_progress";

    const hasTwitchName =
        userInfo?.twitch_name != null &&
        userInfo.twitch_name.trim() !== "";

    const raceEnded = !!raceInfo?.EndedAt || !!raceInfo?.CancelledAt

    const raceLocked = raceInProgress || raceEnded

    const chatRef = useRef<HTMLDivElement | null>(null);

    const myEntrant = raceInfo?.Entrants?.find(
        e => e.user?.id === userInfo?.ID
    );

    const streamBlocksReady = raceStarted && (!myEntrant?.stream_live || myEntrant?.stream_override)

    const canActInRace = joined && raceInProgress && !raceLocked

    const showJoin = !raceLocked
    const disableJoin = raceInProgress || raceEndedOrCancelled || !canJoin || !hasTwitchName
    const canJoinRace = !raceLocked && hasTwitchName && canJoin // server signal
    const showReady = joined && !raceLocked
    const canReady = joined && !raceLocked && !streamBlocksReady
    const showDone = joined && raceStarted
    const canDone = canActInRace
    const showForfeit = joined && raceStarted
    const canForfeit = canActInRace

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
        log.debug("processing unread tab state");

        const messages = raceInfo?.Text ?? [];

        const nextUnread = new Set(unreadTabs);

        for (const message of messages) {
            if (
                message.is_dm &&
                message.user &&
                message.user.id !== activeChatTab
            ) {
                nextUnread.add(message.user.id);

                log.debug(
                    `marked unread dm tab user=${message.user.id}`,
                );
            }
        }

        setUnreadTabs(nextUnread);
    }, [raceInfo?.Text, activeChatTab]);

    const joinDisabledReason =
        !hasTwitchName ? "Link Twitch account" :
            !canJoin ? "Not eligible (stream required or race rules)" :
                raceInProgress ? "Race in progress" :
                    raceEndedOrCancelled ? "Race ended" :
                        "";

    const isAtBottom = () => {
        const el = chatRef.current;

        if (!el) {
            log.debug("chat ref missing while checking scroll position");
            return true;
        }

        return el.scrollHeight - el.scrollTop - el.clientHeight < 50;
    };

    useEffect(() => {
        log.debug("registering chat scroll listener");

        const el = chatRef.current;

        if (!el) {
            log.warn("chat ref unavailable for scroll listener");
            return;
        }

        const onScroll = () => {
            wasAtBottomRef.current =
                el.scrollHeight - el.scrollTop - el.clientHeight < 50;
        };

        el.addEventListener("scroll", onScroll);

        return () => {
            log.debug("removing chat scroll listener");
            el.removeEventListener("scroll", onScroll);
        };
    }, []);

    useLayoutEffect(() => {
        const el = chatRef.current;

        if (!el) {
            return;
        }

        if (wasAtBottomRef.current) {
            log.debug("auto-scrolling chat to bottom");
            el.scrollTop = el.scrollHeight;
        }
    }, [raceInfo?.Text]);

    const handleJoinClick = async () => {
        log.info(`join clicked visible=${joinVisible}`);

        await racetime.Join(joinVisible)

        if (joinVisible) {
            setUserStatus("not_ready");
            log.info("user joined race");
        } else {
            setUserStatus("not_joined");
            log.info("user left race");
        }

        setJoinVisible(!joinVisible)
    }

    const handleReadyClick = async () => {
        log.info(`ready clicked visible=${readyVisible}`);

        await racetime.Ready(readyVisible)

        if (readyVisible) {
            setUserStatus("ready");
            log.info("user marked ready");
        } else {
            setUserStatus("not_ready");
            log.info("user marked unready");
        }

        setReadyVisible(!readyVisible)
    }

    const handleDoneClick = async () => {
        log.info(`done clicked visible=${doneVisible}`);

        await racetime.Done(doneVisible)

        if (doneVisible) {
            setUserStatus("done");
            log.info("user marked done");
        } else {
            setUserStatus("in_progress");
            log.info("user reverted done");
        }

        setDoneVisible(!doneVisible)
    }

    const handleForfeitClick = async () => {
        log.info(`forfeit clicked visible=${forfeitVisible}`);

        await racetime.Forfeit(forfeitVisible)

        if (forfeitVisible) {
            setUserStatus("dnf");
            log.warn("user forfeited");
        } else {
            setUserStatus("in_progress");
            log.info("user unforfeited");
        }

        setForfeitVisible(!forfeitVisible)
    }

    const handleSend = async () => {
        if (!textEntry.trim()) {
            log.warn("attempted to send empty chat message");
            return;
        }

        const id = crypto.randomUUID();

        try {
            log.debug(
                `sending chat message id=${id} length=${textEntry.length}`,
            );

            await racetime.SendText(textEntry, id);

            log.info("chat message sent");

            setTextEntry("");
        } catch (err) {
            log.error("SendText failed", err);
        }
    };

    const handleChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const value = event.target.checked;

        log.info(`hide results changed value=${value}`);

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

    const formatChatTime = (timestamp: string) => {
        const date = new Date(timestamp);

        return date.toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
        });
    };

    const urlRegex = /(https?:\/\/[^\s]+)/g;

    const linkify = (text?: string | null) => {
        if (!text) return text;

        const parts = text.split(urlRegex);

        return parts.map((part, i) => {
            if (part.match(/^https?:\/\//)) {
                return (
                    <a key={i} href={part} target="_blank" rel="noopener noreferrer">
                        {part}
                    </a>
                );
            }
            return part;
        });
    };

    useEffect(() => {
        log.info("subscribing to opensplit connection events");

        return EventsOn("opensplit:connection", (s: ConnectionState) => {
            log.info(
                `opensplit connection updated status=${s.connection_status} message=${s.message}`,
            );

            setOpenSplitConnection(s);
        });
    }, []);

    useEffect(() => {
        if (!raceInfo) return

        if (
            raceStarted &&
            (userStatus === "ready" || userStatus === "not_ready")
        ) {
            log.info(
                `race transitioned to started state userStatus=${userStatus}`,
            );

            if (userStatus === "ready") {
                setUserStatus("in_progress")
            } else {
                if (raceInfo.DisqualifyUnready) {
                    log.warn("user disqualified for not being ready");
                    setUserStatus("dq")
                }
            }
        }
    }, [raceInfo?.StatusVerbose])

    useEffect(() => {
        log.info("subscribing to join eligibility events");

        const eligibilityEvent = EventsOn("joinEligibility", (eligible: boolean) => {
            log.info(`join eligibility updated eligible=${eligible}`);

            setCanJoin(eligible)
        })

        return () => {
            log.debug("unsubscribing from join eligibility events");
            eligibilityEvent()
        }
    }, [])

    useEffect(() => {
        log.info("subscribing to chat update events");

        const newChatText = EventsOn("chatUpdated", (chatText: ChatMessage[]) => {
            log.debug(`chat updated messages=${chatText.length}`);

            const shouldAutoScroll = isAtBottom();

            setRaceInfo((prev) => {
                if (!prev) return prev;

                return { ...prev, Text: chatText };
            });

            wasAtBottomRef.current = shouldAutoScroll;
        });

        return () => {
            log.debug("unsubscribing from chat update events");
            newChatText();
        };
    }, []);

    useEffect(() => {
        log.info("subscribing to user info events");

        const newUserInfo = EventsOn("userInfo", (incoming: UserInfo) => {
            log.info("userInfo event received", incoming);

            setUserInfo(prev => {
                if (!prev) return incoming;
                return {
                    ...prev,
                    ...incoming,
                };
            });
        });

        return () => {
            log.debug("unsubscribing from user info events");
            newUserInfo();
        };
    }, []);

    useEffect(() => {
        log.info("subscribing to race state events");

        const newRaceState = EventsOn("raceStateUpdated", (currentRace: RaceInfo) => {
            log.info(
                `race updated goal=${currentRace.Goal} entrants=${currentRace.EntrantCount}`,
            );

            setRaceInfo(currentRace)
        })

        return () => {
            log.debug("unsubscribing from race state events");
            newRaceState();
        };
    }, []);

    useEffect(() => {
        log.info("subscribing to entrant update events");

        const newEntrants = EventsOn("entrantsUpdated", (entrantList: Entrant[]) => {
            log.debug(`entrants updated count=${entrantList.length}`);

            setEntrantList(entrantList)
        })

        return () => {
            log.debug("unsubscribing from entrant update events");
            newEntrants();
        };
    }, []);

    useEffect(() => {
        log.debug("syncing entrant list into race info");

        setRaceInfo((prev) => {
            if (!prev) return prev

            return {
                ...prev,
                Entrants: entrantList,
            }
        })
    }, [entrantList])

    useEffect(() => {
        log.info("checking stored auth token");

        (
            async () => {
                const raceToken = await racetime.CheckTokens()

                setToken(raceToken)

                log.info(`token check complete present=${raceToken !== ""}`);
            }
        )()
    }, [])

    useEffect(() => {
        if (token == "") {
            log.warn("race polling skipped token missing");
            return
        }

        if (race != "") {
            log.warn(`race polling stopped activeRace=${race}`);
            return
        }

        const fetchRaces = async () => {
            log.debug("fetching race list");

            const raceObj = await RaceList("https://racetime.gg")

            setRaceList(raceObj ?? [])

            log.info(
                `race list updated count=${raceObj?.length ?? 0}`,
            );
        }

        fetchRaces()

        const intervalId = setInterval(() => {
            fetchRaces()
        }, 5000)

        return () => {
            log.debug("stopping race polling interval");
            clearInterval(intervalId)
        }
    }, [token, race])

    useEffect(() => {
        if (race !== "") {
            WindowSetSize(900, 700);
        } else {
            WindowSetSize(320, 580);
        }
    }, [race]);

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
                                LogInfo(`joining websocket race=${item.URL}`);

                                setJoinedRace(item.URL);

                                await racetime.WebSocketConnection(item.URL);

                                LogInfo(`websocket connected race=${item.URL}`);
                            } catch (err) {
                                LogError(`failed to connect websocket: ${err}`);
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
                    <div className="raceHeader">
                        {/* LEFT SIDE (EMPTY / aligns with left column) */}
                        {/* <div className="raceHeaderLeft" /> */}

                        {/* RIGHT SIDE (Back + Connection Status) */}
                        {/* <div className="raceHeaderRight"> */}
                        {/* </div> */}
                    </div>
                    <div className="raceMain">

                        <div className="raceLeft">
                            <div className="raceInfoBlock">

                                <div className="raceInfoRow">
                                    <span className="label">Game:</span>
                                    <span className="value">{raceInfo?.Game}</span>
                                </div>

                                <div className="raceInfoRow">
                                    <span className="label">Race:</span>
                                    <span className="value">{race}</span>
                                </div>

                                <div className="raceInfoRow">
                                    <span className="label">Goal:</span>
                                    <span className="value">{raceInfo?.Goal}</span>
                                </div>

                                <div className="raceInfoRow">
                                    <span className="label">Info:</span>
                                    <div className="value">
                                        {raceInfo?.Info ? linkify(raceInfo.Info) : ""}
                                    </div>
                                </div>

                            </div>

                            <div className="chatContainer">

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

                                <div
                                    ref={chatRef}
                                    className="chatBox">

                                    {filteredMessages.map((message) => {
                                        const senderName = message.is_bot
                                            ? (message.bot || "Bot")
                                            : (message.user?.name ?? "System");

                                        return (
                                            <div
                                                key={message.id}
                                                className={
                                                    message.is_dm
                                                        ? "dmMessage"
                                                        : "mainMessage"
                                                }>

                                                <div className="chatText">
                                                    <span className="chatTimestamp">
                                                        {formatChatTime(message.posted_at)}
                                                    </span>
                                                    {" "}
                                                    <span className="chatSender">
                                                        {senderName}:
                                                    </span>
                                                    {" "}
                                                    {linkify(message.message)}
                                                </div>

                                            </div>
                                        );
                                    })}
                                </div>
                            </div>
                        </div>

                        <div className="entrantPanel">
                            <div className="raceStatusPanel">
                                <button
                                    className="backButton"
                                    onClick={async () => {
                                        LogInfo(`disconnecting from race`);
                                        await racetime.Join(false);
                                        await racetime.DisconnectRace();

                                        setJoinVisible(true);
                                        setReadyVisible(true);
                                        setDoneVisible(true);
                                        setForfeitVisible(true);

                                        setUserStatus("not_joined");

                                        setJoinedRace("");
                                        setRaceInfo(undefined);
                                        setEntrantList([]);
                                    }}
                                >
                                    Back to Races
                                </button>

                                <div className="status">
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
                                                    />
                                                </td>
                                                <td>{openSplitConnection.message}</td>
                                            </tr>
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                            <div className="raceStatusPanel">
                                <div className="timerDisplay">
                                    {raceInfo?.StatusVerbose}
                                </div>

                                <div className="timerStatus">
                                    {raceInfo?.StatusHelpText}
                                </div>
                                <div>
                                    {"Ranked: " + (raceInfo?.Ranked ? "Yes" : "No")}
                                </div>

                                <div>
                                    {"Auto Start: " + (raceInfo?.AutoStart ? "Enabled" : "Disabled")}
                                </div>
                            </div>

                            <div className="entrantList">
                                {raceInfo?.Entrants?.map((entrant, index) => (
                                    <div
                                        key={index}
                                        className="entrantRow">

                                        <img
                                            src={
                                                entrant.stream_live ||
                                                    entrant.stream_override
                                                    ? connected
                                                    : disconnected
                                            }
                                            alt="stream"
                                            width={16}
                                            height={16}
                                        />

                                        <img
                                            src={entrant.user.avatar}
                                            alt={entrant.user.name}
                                            width={24}
                                            height={24}
                                        />

                                        <span>{entrant.place_ordinal}</span>
                                        <span>{entrant.user.name}</span>
                                        <span>{entrant.value}</span>
                                    </div>
                                ))}

                                <div className="entrantSummary">
                                    {raceInfo?.EntrantCount} entrants (
                                    {raceInfo?.EntrantInactiveCount} inactive)
                                </div>
                            </div>

                        </div>
                    </div>

                    <div className="actionPanel">

                        <label>
                            <input
                                type="checkbox"
                                onChange={handleChange}
                            />
                            Hide Results
                        </label>

                        <button
                            onClick={async () => {
                                await racetime.SaveLog();
                            }}>
                            Save Log
                        </button>

                        <button
                            disabled={!canJoinRace}
                            hidden={!showJoin}
                            onClick={handleJoinClick}
                        >
                            {joinVisible ? "Join" : "Leave"}
                        </button>
                        {disableJoin && (
                            <div className="hint">
                                Cannot join: {joinDisabledReason}
                            </div>
                        )}
                        <button
                            disabled={!canReady}
                            hidden={!showReady}
                            onClick={handleReadyClick}
                        >
                            {readyVisible ? "Ready" : "Unready"}
                        </button>

                        <button
                            disabled={!canDone}
                            hidden={!showDone}
                            onClick={handleDoneClick}
                        >
                            {!doneVisible ? "Done" : "Undone"}
                        </button>

                        <button
                            disabled={!canForfeit}
                            hidden={!showForfeit}
                            onClick={handleForfeitClick}
                        >
                            {!forfeitVisible ? "Forfeit" : "Unforfeit"}
                        </button>

                    </div>

                    <div className="chatInputBar">

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

                        <button onClick={handleSend}>
                            Send
                        </button>

                    </div>
                </div>
            )
        }
    }
}

export default App
