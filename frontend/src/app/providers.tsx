import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { App, ConfigProvider, theme as antdTheme } from "antd";
import type { PropsWithChildren } from "react";
import { ThemeProvider, useTheme } from "../shared/adapters/theme";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 15_000
    }
  }
});

const cssVar = (name: string) => `var(${name})`;

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <ThemeProvider>
      <AntdThemeProvider>{children}</AntdThemeProvider>
    </ThemeProvider>
  );
}

function AntdThemeProvider({ children }: PropsWithChildren) {
  const { resolvedTheme } = useTheme();
  const isDark = resolvedTheme === "dark";

  return (
    <ConfigProvider
      theme={{
        algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
        cssVar: {},
        token: {
          colorPrimary: cssVar("--primary"),
          colorInfo: cssVar("--secondary"),
          colorSuccess: cssVar("--chart-4"),
          colorWarning: cssVar("--chart-3"),
          colorError: cssVar("--destructive"),
          colorText: cssVar("--foreground"),
          colorTextSecondary: cssVar("--muted-foreground"),
          colorTextTertiary: cssVar("--muted-foreground"),
          colorBgBase: cssVar("--background"),
          colorBgContainer: cssVar("--card"),
          colorBgElevated: cssVar("--popover"),
          colorBorder: cssVar("--border"),
          colorBorderSecondary: cssVar("--border"),
          colorFillAlter: cssVar("--muted"),
          borderRadius: 12,
          fontFamily: "Geist, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif"
        },
        components: {
          Button: {
            borderRadius: 999,
            controlHeightLG: 44
          },
          Card: {
            borderRadiusLG: 16
          },
          Menu: {
            itemBorderRadius: 10
          },
          Table: {
            headerBg: cssVar("--muted"),
            rowHoverBg: cssVar("--accent")
          }
        }
      }}
    >
      <App>
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
      </App>
    </ConfigProvider>
  );
}
