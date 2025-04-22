import { parseArgs } from "node:util";
import { readFunctionDefinition } from "../lib/utils";
import { runDockerBenchmark } from "../lib/benchmark";

const { positionals } = parseArgs({
	options: {},
	allowPositionals: true,
});

const [funcName] = positionals;
console.log(`Benchmarking ${funcName}`);

console.log("loading definition...");
const definition = await readFunctionDefinition(`./targets/${funcName}.json`);
console.log(definition);

const abortController = new AbortController();
process.on("SIGINT", () => {
	abortController.abort();
});

await runDockerBenchmark(definition, {
	baseDir: "./",
	baseTestId: new Date().toISOString(),
	abortController,
});
