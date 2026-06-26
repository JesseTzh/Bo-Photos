import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { App, ConfigProvider } from "antd";
import type { PropsWithChildren } from "react";
import { ThemeProvider } from "../shared/adapters/theme";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 15_000
    }
  }
});

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <ThemeProvider>
      <ConfigProvider
        theme={{
          cssVar: {},
          token: {
            colorPrimary: "#d97706",
            colorInfo: "#0891b2",
            colorSuccess: "#10b981",
            colorWarning: "#f59e0b",
            colorError: "#ef4444",
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
            }
          }
        }}
      >
        <App>
          <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
        </App>
      </ConfigProvider>
    </ThemeProvider>
  );
}
