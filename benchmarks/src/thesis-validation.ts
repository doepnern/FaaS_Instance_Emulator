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
  console.log(`Validating ${funcName}`);
  for (let i = 1; i < 2; i++) {
    if (abortController.signal.aborted) break;
    const definition = await readFunctionDefinition(`./targets/${funcName}-validation.json`);
    definition.name = `${funcName}-validation-${i}`;
    definition.benchmarks = definition.benchmarks.map(b => ({
      ...b, type: "validation", options: {
        ...b.options,
        duration: "10m",
      }
    }));
    console.log(`Running ${definition.name}`);
    try {
      await runDockerBenchmark(definition, {
        baseDir: "./",
        baseTestId: new Date().toISOString(),
        abortController,
      });
    } catch (e) {
      console.error(e);
    }
  }
}




