import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { execFileSync } from "node:child_process";
import { defineConfig } from "vite";
import { fileURLToPath, URL } from "node:url";

function createBuildVersion() {
  const now = new Date();
  const dateParts = Object.fromEntries(
    new Intl.DateTimeFormat("en-CA", {
      timeZone: "Asia/Shanghai",
      year: "2-digit",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      hourCycle: "h23"
    })
      .formatToParts(now)
      .map(({ type, value }) => [type, value])
  );
  const timestamp = [
    dateParts.year,
    dateParts.month,
    dateParts.day,
    dateParts.hour,
    dateParts.minute
  ].join("-");

  let gitSha = process.env.BOPHOTOS_GIT_SHA?.trim().slice(0, 7) || "";
  if (!gitSha) {
    try {
      gitSha = execFileSync("git", ["rev-parse", "--short=7", "HEAD"], {
        cwd: fileURLToPath(new URL("..", import.meta.url)),
        encoding: "utf8"
      }).trim();
    } catch {
      gitSha = "unknown";
    }
  }

  return `${timestamp}-${gitSha}`;
}

export default defineConfig({
  define: {
    __APP_VERSION__: JSON.stringify(createBuildVersion())
  },
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "~": fileURLToPath(new URL("./src", import.meta.url)),
      "@": fileURLToPath(new URL("./src", import.meta.url))
    }
  },
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8080",
        changeOrigin: false
      },
      "/health": {
        target: "http://127.0.0.1:8080",
        changeOrigin: false
      },
      "/media": {
        target: "http://127.0.0.1:8080",
        changeOrigin: false
      }
    }
  },
  build: {
    outDir: "dist",
    sourcemap: true,
    chunkSizeWarningLimit: 900,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes("node_modules")) return;
          if (/[\\/]node_modules[\\/](react|react-dom|react-router-dom)[\\/]/.test(id)) return "react";
          if (id.includes("@tanstack/react-query")) return "query";
          if (id.includes("@ant-design/icons") || id.includes("lucide-react") || id.includes("@radix-ui/react-icons")) return "icons";
          if (id.includes("framer-motion") || id.includes("motion-dom") || id.includes("motion-utils")) return "motion";
          if (id.includes("antd")) return "antd";
        }
      }
    }
  }
});
