import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    host: "0.0.0.0",
    proxy: {
      "/api": {
        target: "http://127.0.0.1:9999",
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: "dist",
    assetsDir: "assets",
    sourcemap: true
  }
});
