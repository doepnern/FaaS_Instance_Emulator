import type { ContainerInfo } from "../utils";
import { DockerStatsSchema, type DockerStatsMetric } from "./schema";

export class DockerMetricsCollector {
	dockerContainers: ContainerInfo[];
	socketPath: string;
	abortController?: AbortController;
	cancelPromise?: Promise<unknown>;

	constructor(dockerContainerName: ContainerInfo[], socketPath: string) {
		this.dockerContainers = dockerContainerName;
		this.socketPath = socketPath;
	}
	async start(
		onProgress: (data: DockerStatsMetric, conainerInfo: ContainerInfo) => void,
	) {
		const ac = new AbortController();
		this.abortController = ac;
		const cancelPromises = await Promise.all(
			this.dockerContainers.map(async (cn) => {
				const statsRes = await fetch(
					`http://${this.socketPath}/containers/${cn.name}/stats`,
				);
				if (!statsRes.body)
					throw new Error("No response fromd docker for stats");

				const cancelPromise = handleStreamProgress(
					statsRes.body,
					ac.signal,
					(metric) => onProgress(metric, cn),
				);
				return () => cancelPromise;
			}),
		);

		this.cancelPromise = Promise.all(cancelPromises.map((cp) => cp()));

		return this.abortController;
	}
	async stop() {
		if (this.abortController) this.abortController.abort();
		return this.cancelPromise;
	}
}

async function handleStreamProgress(
	stream: ReadableStream,
	signal: AbortSignal,
	onProgress: (data: DockerStatsMetric) => void,
) {
	const decoder = new TextDecoder();
	let textBuffer = "";
	streamLoop: for await (const chunk of stream) {
		textBuffer += decoder.decode(chunk);
		if (textBuffer.length === 0) continue;
		const parts = textBuffer.split("\n");
		if (parts.length < 2) continue;
		try {
			for (const p of parts.slice(0, -1)) {
				if (p.length === 0) continue;
				const metric = DockerStatsSchema.safeParse(JSON.parse(p));
				if (!metric.success) {
					console.error(p);
					console.error(metric.error);
					throw new Error("Received invalid metric from Docker api");
				}
				if (signal.aborted) {
					// this is needed because stram.cancel can only be called outside the Stream consuming loop
					break streamLoop;
				}
				onProgress(metric.data);
			}
			textBuffer = parts[parts.length - 1];
		} catch (err) {
			console.error(textBuffer);
			console.error(err);
			textBuffer = "";
		}
	}
	await stream.cancel();
}
