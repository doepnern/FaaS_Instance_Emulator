import { sleep, $ } from "bun";
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
  console.log(`Validating ${funcName}`);
  for (let i = 1; i < 2; i++) {
    if (abortController.signal.aborted) break;
    const p = Bun.spawn(['./faas-simulator', funcName], {
      cwd: "../simulator-linked-list",
      onExit: () => {
        console.log(` ${funcName} exited successsfully`);
      }
    });

    const definition = await readFunctionDefinition(`./targets/${funcName}-validation.json`);
    definition.name = `${funcName}-mock-validation-${i}`;
    definition.benchmarks = definition.benchmarks.map(b => ({
      ...b, type: "validation-mock-server", options: {
        ...b.options,
        duration: "10m",
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
      console.log(`Killing ${funcName}`);
      p.kill();
      await p.exited
      console.log(`Killed ${funcName}`);
      await sleep(1000);
    }
  }
}




