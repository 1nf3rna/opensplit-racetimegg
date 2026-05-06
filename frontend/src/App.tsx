import { useEffect, useState } from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import * as racetime from "../wailsjs/go/main/App";
import { LoginWithOAuth, RaceList, RaceWindow } from './components/racetime_gg';
import { WindowSetSize } from "../wailsjs/runtime";
import ButtonList, { ButtonData } from "./components/ButtonList"

function App() {
    const [token, setToken] = useState("")
    const [raceList, setRaceList] = useState<ButtonData[]>([])
    const [race, setJoinedRace] = useState("")

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
            <div id="App">
                <button
                    onClick={async () => {
                        await LoginWithOAuth()
                        // This just triggers the useeffects functions
                        setToken("get tokens")
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
                    <ButtonList
                        data={raceList}
                        onClick={(item) => {
                            console.log("Clicked", item);
                            setJoinedRace(item.URL)
                        }}
                    />
                </div>
            )
        } else {
            // race selected
            // show race window
            return (
                <div id="App">
                    {/* {/* <ul> */}
                        {/* { */}
                            {/* // raceList.map((r) => <li>{r["name"]}</li>) */}
                            RaceWindow(w, race)
                        {/* // } */}
                    {/* </ul> */}
                </div>
            )
        }
    }
}

export default App
