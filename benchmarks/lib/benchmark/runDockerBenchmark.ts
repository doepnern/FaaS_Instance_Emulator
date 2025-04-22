import { $, sleep } from "bun";
import {
	BASE_PORT,
	HOST,
	startContainers,
	stopContainers,
	type FunctionDefinition,
} from "../utils";
import { getDockerInfo } from "../dockerApi/info";
import path from "node:path";
import type { MultiRpsEnv } from "../types";
import { collectMetrics } from "./collectMetrics";
import { readFileSync, writeFileSync } from "node:fs";
import { fetchDockerApi } from "../dockerApi/conifg";

export async function runDockerBenchmark(
	definition: FunctionDefinition,
	{
		baseDir = "",
		baseTestId,
		abortController,
		infuxIntegration = false,
	}: {
		baseDir: string;
		baseTestId: string;
		abortController?: AbortController;
		infuxIntegration?: boolean;
	},
) {
	const shell = $.cwd(baseDir);

	const { info, version } = await getDockerInfo();
	console.log(`Detected docker version ${info.ServerVersion}`);
	console.log(`Detected api version ${version.ApiVersion}`);

	try {
		await shell`mkdir -p ./results`;
		await shell`rm ./results/${definition.name}/*`.nothrow();
		await shell`mkdir -p ./results/${definition.name}`;

		const conatinerInfoRes = await fetchDockerApi(
			`/containers/${definition.docker_image_name}/json`,
		);
		const containerInfo = await conatinerInfoRes.json();
		console.log(
			`Located container ${definition.docker_image_name} (${containerInfo.Id})`,
		);

		const benchmarks = definition.benchmarks;

		console.log(
			`Running benchmarks: ${benchmarks.map((b) => b.name).join(", ")}`,
		);

		for (const benchmark of benchmarks) {
			const startedAt = new Date();
			const containers = await startContainers({
				...definition,
				replicas: benchmark?.replicas ?? definition.replicas,
			});
			await sleep(5000);

			if (abortController?.signal.aborted) break;
			const testId = `${benchmark.name}_${baseTestId}`;

			const k6Env: MultiRpsEnv = {
				...process.env,
				FUNC_NAME: definition.name,
				PAYLOAD: JSON.stringify(definition.payload),
				RESPONSE: definition.response
					? JSON.stringify(definition.response)
					: "",
				HOST,
				PORT_RANGE: `${BASE_PORT}-${BASE_PORT + ((benchmark.replicas ?? definition.replicas) - 1)}`,
				PATH: definition.path,
				K6_WEB_DASHBOARD: "true",
				BENCHMARK_TEST_NAME: benchmark.name,
				BENCHMARK: JSON.stringify(benchmark),
				REPLICAS: `${benchmark.replicas ?? definition.replicas}`,
				IMAGE_NAME: definition.docker_image_name,
			};

			const k6MetricsFile = `./results/${definition.name}/${benchmark.name}_metrics.json`;
			const prometheusMetricsFile = `./results/${definition.name}/${benchmark.name}_metrics_prometheus.json`;

			const proc = Bun.spawn(
				[
					"./k6",
					"run",
					`./k6Scripts/${benchmark.type}.js`,
					"--tag",
					`testid=${benchmark.name}_${testId}`,
					"--out",
					`json=${k6MetricsFile}`,
					...(infuxIntegration
						? ["--out", "influxdb=http://localhost:8086"]
						: []),
				],
				{
					env: k6Env,
					stdout: "inherit",
					cwd: baseDir,
				},
			);

			abortController?.signal.addEventListener("abort", () => {
				proc.kill();
			});

			await proc.exited;
			const endedAt = new Date();

			console.log("exited");

			await stopContainers({ containers: containers });
			await sleep(30000);
			const metrics = await collectMetrics({
				from: startedAt,
				until: endedAt,
			});
			await writeFileSync(prometheusMetricsFile, JSON.stringify(metrics));
			const summary = JSON.parse(
				readFileSync(
					`./results/${definition.name}/${benchmark.name}_summary.json`,
					{ encoding: "utf-8" },
				),
			);
			await writeFileSync(
				`./results/${definition.name}/${benchmark.name}_result.json`,
				JSON.stringify({
					startedAt,
					endedAt,
					k6MetricsFile: path.resolve(k6MetricsFile),
					prometheusMetricsFile: path.resolve(prometheusMetricsFile),
					dockerMetricsFile: path.resolve(
						`./results/${definition.name}/${benchmark.name}_docker-metrics`,
					),
					summary,
				}),
			);
			await sleep(12000);
		}

		console.log("cleaning up...");
	} catch (err) {
		console.error(err);
	}
}
