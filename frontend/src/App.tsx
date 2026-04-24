import { useEffect, useState } from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import * as racetime from "../wailsjs/go/main/App";
import { LoginWithOAuth, RaceList/*, RaceListWindow*/ } from './components/racetime_gg';
import { WindowSetSize } from "../wailsjs/runtime";

function App() {
    const [token, setToken] = useState("")
    const [raceList, setRaceList] = useState([])
    const [race, setJoinedRace] = useState("")

    // Gets tokens from backend
    useEffect(() => {
        // call backend function to get token
        async () => {
            const raceToken = await racetime.CheckTokens()
            setToken(raceToken)
        }
    }, [])

    // Gets list of races
    useEffect(() => {
        if (token == "") {
            return
        }

        (
            async () => {
                const raceObj = await RaceList()
                setRaceList(raceObj["races"])
            }
        )()
    }, [token])

    // Gets selected race
    // useEffect(() => {
    //     if (race == "") {
    //         return
    //     }

    //     (
    //         async () => {
    //             const selectedRace = await GetRace()
    //             setJoinedRace(selectedRace)
    //         }
    //     )()
    // }, [])

    WindowSetSize(320, 580);
    if (token == "") {
        // no token
        // show login button
        return (
            <div id="App">
                <button
                    onClick={async () => {
                        await LoginWithOAuth();
                    }}
                >
                    Racetime.gg Auth
                </button>
            </div>
        )
    } else {
        if (race == "") {
            // no race
            // show race list buttons
            return (
                <div id="App">
                    <ul>
                        {
                            raceList.map((r) => <li>{r["name"]}</li>)
                        }
                    </ul>
                </div>
            )
        } else {
            // race selected
            // show race window
            return (
                <div id="App">
                    {/* <ul>
                        {
                            raceList.map((r) => <li>{r["name"]}</li>)
                        }
                    </ul> */}
                </div>
            )
        }
    }
}

export default App
