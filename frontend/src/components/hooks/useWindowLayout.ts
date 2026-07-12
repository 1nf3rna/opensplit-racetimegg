import { useEffect } from "react";
import { WindowSetSize } from "../../../wailsjs/runtime";

export function useWindowLayout(race: string) {
  useEffect(() => {
    if (race) {
      WindowSetSize(900, 700);
    } else {
      WindowSetSize(320, 580);
    }
  }, [race]);
}
