import { runDockerBenchmark } from "../lib/benchmark";
import { readFunctionDefinition } from "../lib/utils";

const functions = [
  "go-io",
  "go-factorial",
  "go-string-reversal"
]

const abortController = new AbortController();
process.on("SIGINT", () => {
  abortController.abort();
});

for (const funcName of functions) {
  console.log(`Benchmarking ${funcName}`);
  for (let i = 1; i < 4; i++) {
    if (abortController.signal.aborted) break;
    const definition = await readFunctionDefinition(`./targets/${funcName}-breakpoint.json`);
    definition.name = `${funcName}-breakpoint-${i}`;
    definition.benchmarks = definition.benchmarks.map(b => ({
      ...b, type: "breakpoint", options: {
        ...b.options,
        duration: "20m",
      }
    }));
    console.log(`Running ${definition.name}`);
    await runDockerBenchmark(definition, {
      baseDir: "./",
      baseTestId: new Date().toISOString(),
      abortController,
    });
  }
}




