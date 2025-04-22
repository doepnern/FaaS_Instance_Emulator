import { $ } from "bun";
import { runMockServerBenchmark } from "../lib/benchmark";
import { readFunctionDefinition } from "../lib/utils";

const functions = [
  "go-factorial",
  "go-string-reversal",
  "go-io",
]

const abortController = new AbortController();
process.on("SIGINT", () => {
  abortController.abort();
});

await $`go build`.cwd("../simulator-linked-list").env({ GOPATH: `${process.env.HOME}/go`, HOME: process.env.HOME! });

for (const funcName of functions) {
  console.log(`Benchmarking ${funcName}`);
  for (let i = 1; i < 2; i++) {
    if (abortController.signal.aborted) break;
    const p = Bun.spawn(['./faas-simulator', funcName], {
      cwd: "../simulator-linked-list",
      onExit: () => {
        console.log(` ${funcName} exited successsfully`);
      }
    });

    const definition = await readFunctionDefinition(`./targets/${funcName}-breakpoint.json`);
    definition.name = `${funcName}-breakpoint-mock-${i}`;
    definition.benchmarks = definition.benchmarks.map(b => ({
      ...b, type: "breakpoint-mock-server", options: {
        ...b.options,
        duration: "20m",
      }
    }));
    console.log(`Running ${definition.name}`);
    try {
      await runMockServerBenchmark(definition, {
        baseDir: "./",
        baseTestId: new Date().toISOString(),
        abortController,
      });
    } catch (e) {
      console.error(e);
    } finally {
      p.kill();
      await p.exited
    }
  }
}

console.log("Done");
process.exit(0);



