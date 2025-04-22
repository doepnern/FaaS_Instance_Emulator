import { parseArgs } from "node:util";
import { readFunctionDefinition } from "../lib/utils";
import { runMockServerBenchmark } from "../lib/benchmark";

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

const p = Bun.spawn(['go' , 'run', '.', definition.docker_image_name], {
	cwd: "../simulator-linked-list",
	// stdout: outlog,
});

await runMockServerBenchmark(definition, {
	baseTestId: new Date().toISOString(),
	abortController,
});
console.log("killing simulator");
p.kill();
await p.exited;
console.log("simulator killed");
process.exit(0);
