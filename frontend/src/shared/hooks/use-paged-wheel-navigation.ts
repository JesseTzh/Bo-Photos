import { type RefObject, useEffect, useRef } from "react";

export function usePagedWheelNavigation(
  target: RefObject<HTMLElement | null>,
  onNavigate: (direction: -1 | 1) => void,
  enabled = true
) {
  const callback = useRef(onNavigate);

  useEffect(() => {
    callback.current = onNavigate;
  }, [onNavigate]);

  useEffect(() => {
    const element = target.current;
    if (!element || !enabled) return;

    let accumulatedDelta = 0;
    let locked = false;
    let resetTimer: number | undefined;
    let unlockTimer: number | undefined;

    const handleWheel = (event: WheelEvent) => {
      if (event.ctrlKey || event.deltaY === 0) return;
      event.preventDefault();
      if (locked) return;

      const multiplier = event.deltaMode === WheelEvent.DOM_DELTA_LINE
        ? 16
        : event.deltaMode === WheelEvent.DOM_DELTA_PAGE
          ? window.innerHeight
          : 1;
      accumulatedDelta += event.deltaY * multiplier;
      window.clearTimeout(resetTimer);
      resetTimer = window.setTimeout(() => {
        accumulatedDelta = 0;
      }, 140);

      if (Math.abs(accumulatedDelta) < 36) return;
      const direction = accumulatedDelta > 0 ? 1 : -1;
      accumulatedDelta = 0;
      locked = true;
      callback.current(direction);
      unlockTimer = window.setTimeout(() => {
        locked = false;
      }, 700);
    };

    element.addEventListener("wheel", handleWheel, { passive: false });
    return () => {
      element.removeEventListener("wheel", handleWheel);
      window.clearTimeout(resetTimer);
      window.clearTimeout(unlockTimer);
    };
  });
}
