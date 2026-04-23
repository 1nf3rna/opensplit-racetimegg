import { useEffect, useState } from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import * as racetime from "../wailsjs/go/main/App";
import { LoginWithOAuth, RaceList, RaceListWindow } from './components/racetime_gg';
import { WindowSetSize } from "../wailsjs/runtime";

function App() {
    const [token, setToken] = useState("")
    const [raceList, setRaceList] = useState([])
    useEffect(() => {
        // call backend function to get token
        // setToken to return
        // if (await CheckTokens()) {
        // return
        // }
        async () => {
                const raceToken = await racetime.GetAccessToken()
                setToken(raceToken)
            }
    }, [])
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

    WindowSetSize(320, 580);
    if (token == "") {
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
    }
    return (
        <div id="App">
            <ul>
            {
                raceList.map((r)=><li>{r["name"]}</li>)
            }
            </ul>
        </div>
    )
}

export default App
