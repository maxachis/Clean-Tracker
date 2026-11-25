// vite build empties dist/, but go:embed in main.go needs it to hold a file
// on a fresh clone (dist/ is otherwise gitignored), so restore the keep file.
import { writeFileSync } from "node:fs";

writeFileSync(new URL("../dist/.gitkeep", import.meta.url), "");
