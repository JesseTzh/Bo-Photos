import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { execFileSync } from "node:child_process";
import { defineConfig } from "vite";
import { fileURLToPath, URL } from "node:url";

function createBuildVersion() {
  const now = new Date();
  const pad = (value: number) => String(value).padStart(2, "0");
  const timestamp = [
    pad(now.getFullYear() % 100),
    pad(now.getMonth() + 1),
    pad(now.getDate()),
    pad(now.getHours()),
    pad(now.getMinutes())
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
