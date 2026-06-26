import { createContext, useCallback, useContext, useEffect, useMemo, useState, type PropsWithChildren } from "react";

type ThemeMode = "light" | "dark";

interface ThemeContextValue {
  resolvedTheme: ThemeMode;
  setTheme: (theme: ThemeMode) => void;
  toggle: () => void;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

function readInitialTheme(): ThemeMode {
  if (typeof window === "undefined") return "light";
  const stored = window.localStorage.getItem("bophotos-theme");
  if (stored === "dark" || stored === "light") return stored;
  return window.matchMedia?.("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

export function ThemeProvider({ children }: PropsWithChildren) {
  const [theme, setThemeState] = useState<ThemeMode>(() => readInitialTheme());

  const setTheme = useCallback((next: ThemeMode) => {
    setThemeState(next);
    window.localStorage.setItem("bophotos-theme", next);
  }, []);

  const toggle = useCallback(() => {
    setThemeState((value) => {
      const next = value === "dark" ? "light" : "dark";
      window.localStorage.setItem("bophotos-theme", next);
      return next;
    });
  }, []);

  useEffect(() => {
    document.documentElement.classList.toggle("dark", theme === "dark");
    document.documentElement.dataset.theme = theme;
  }, [theme]);

  const value = useMemo(() => ({ resolvedTheme: theme, setTheme, toggle }), [setTheme, theme, toggle]);
  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme() {
  const value = useContext(ThemeContext);
  if (!value) throw new Error("useTheme must be used inside ThemeProvider");
  return value;
}
