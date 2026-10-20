import {
  RestartContainer,
  Scan,
  StartContainer,
  StopContainer,
} from "../../wailsjs/go/main/App.js";
import type { main } from "../../wailsjs/go/models";

type ContainerAction = "start" | "stop" | "restart";

const actions: Record<ContainerAction, (id: string) => Promise<void>> = {
  start: StartContainer,
  stop: StopContainer,
  restart: RestartContainer,
};

// Wails rejects promises with a plain string, not an Error.
function message(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

// One shared scan state for every view: a single source keeps the lists, the
// totals and the cleanup report consistent with each other.
class ScanState {
  result = $state<main.ScanResult | undefined>();
  loading = $state(false);
  error = $state("");
  // ID of the container whose action is running, so only its row is disabled.
  busy = $state("");

  async refresh(): Promise<void> {
    this.loading = true;
    try {
      this.result = await Scan();
      this.error = "";
    } catch (err) {
      this.error = message(err);
    } finally {
      this.loading = false;
    }
  }

  async containerAction(action: ContainerAction, id: string): Promise<void> {
    this.busy = id;
    try {
      await actions[action](id);
    } catch (err) {
      this.error = message(err);
    } finally {
      this.busy = "";
    }
    await this.refresh();
  }
}

export const scan = new ScanState();
