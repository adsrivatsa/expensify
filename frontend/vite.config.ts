import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      // Proxy /auth and /api to the Go backend so cookies work on the same origin.
      "/auth": {
        target: "https://expensify-backend.adsrivatsa.com",
        changeOrigin: true,
      },
      "/api": {
        target: "https://expensify-backend.adsrivatsa.com",
        changeOrigin: true,
      },
    },
    allowedHosts: ["expensify.adsrivatsa.com"],
  },
});
