import { writeFileSync } from "node:fs";
import { defineConfig } from "vite";

// Vite empties dist/ on every build. The Go side embeds it, so `go vet` and `go test` need the
// directory to exist even when nobody has built the frontend; keep the committed placeholder.
const keepPlaceholder = () => ({
  name: "keep-dist-placeholder",
  closeBundle() {
    writeFileSync("dist/.gitkeep", "");
  },
});

export default defineConfig({
  // VITE_HOST=0.0.0.0 in the dev container, so the editor's port forward can reach the server.
  server: { host: process.env.VITE_HOST || "127.0.0.1", port: Number(process.env.WAILS_VITE_PORT) || 9245, strictPort: true },
  plugins: [keepPlaceholder()],
});
