import {useState} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import * as racetime from "../wailsjs/go/main/App";
import { LoginWithOAuth, RaceListWindow } from './components/racetime_gg';
import { WindowSetSize } from "../wailsjs/runtime";

function App() {
    WindowSetSize(320, 580);
    return (
        <div id="App">
            <button
                onClick={async () => {
                    await LoginWithOAuth();
                }}
            >
                Racetime.gg Auth
            </button>

            {/* <button
                hidden
                onClick={async () => {
                    // TODO: Get this shit to work
                    await RaceListWindow();
                }}
            >
                Racetime.gg Races
            </button> */}
        </div>
    )
}

export default App
