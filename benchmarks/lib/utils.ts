import { sleep } from "bun";
import { $ as baseShell } from "bun";
import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import { z } from "zod";

export const BASE_PORT = 5781;
export const DOCKER_HOST = import.meta.env.DOCKER_HOST;
export const HOST = import.meta.env.HOST ?? "http://localhost";

export const DOCKER_SOCKET_PATH = "/var/run/docker.sock";
const $ = baseShell.env({ DOCKER_HOST });

const BenchmarkDefinitionSchma = z.object({
	name: z.string().min(1).default("benchmark"),
	type: z.enum(["breakpoint", "breakpoint-mock-server", "validation", "validation-mock-server"]).default("breakpoint"),
	replicas: z.number().optional(),
	options: z.record(z.any()).optional().default({}),
});

export const FunctionDefinitionSchema = z.object({
	name: z.string().min(1),
	payload: z.any(),
	path: z.string().default(""),
	response: z.any().optional(),
	docker_image_name: z.string(),
	benchmarks: z.array(BenchmarkDefinitionSchma),
	replicas: z.coerce.number().default(1),
	docker_image_port: z.coerce.number(),
	constraints: z
		.object({
			cpus: z.coerce.number().min(0.1),
			memory: z.string().min(1),
		})
		.partial()
		.optional(),
});

export type FunctionDefinition = z.infer<typeof FunctionDefinitionSchema>;

export async function readFunctionDefinition(path: string) {
	const stringData = await readFile(path, {
		encoding: "utf8",
	});
	const definition = FunctionDefinitionSchema.parse(JSON.parse(stringData));
	return definition;
}

function constraintsToDockerRunOptions(
	constraints: FunctionDefinition["constraints"],
) {
	return constraints
		? Object.entries(constraints)
			.map(([key, val]) => `--${key}="${val}"`)
			.join(" ")
		: "";
}

export async function readFunctionDefinitionsInFolder(folderPath: string) {
	const allFiles = await readdir(folderPath);
	return Promise.all(
		allFiles.map((fileName) =>
			readFunctionDefinition(path.join(folderPath, fileName)),
		),
	);
}

export type ContainerInfo = {
	port: number;
	name: string;
	constraints: FunctionDefinition["constraints"];
};

export async function startContainers({
	docker_image_name,
	docker_image_port,
	replicas,
	constraints,
}: Pick<
	FunctionDefinition,
	"docker_image_name" | "docker_image_port" | "replicas" | "constraints"
>) {
	const containerInfo: ContainerInfo[] = Array.from({ length: replicas }).map(
		(_, ind) => ({
			port: BASE_PORT + ind,
			name: `${docker_image_name}_${ind}`,
			constraints,
		}),
	);

	console.log("removing potentially orphaned containers from previous runs...");
	await stopContainers({ containers: containerInfo });

	return Promise.all(
		containerInfo.map(async (info) => {
			console.log(`Starting container ${info.name}`);
			// CAVE: This is not safe
			await $`docker run -d --restart unless-stopped --name ${info.name} ${{ raw: constraintsToDockerRunOptions(constraints) }} -p ${info.port}:${docker_image_port} ${docker_image_name}`;
			return info;
		}),
	);
}

export async function stopContainers({
	containers,
}: { containers: ContainerInfo[] }) {
	await Promise.allSettled(
		containers.map(async (c) => $`docker stop ${c.name}`),
	);
	await sleep(5000);
	return Promise.allSettled(
		containers.map(async (c) => $`docker rm ${c.name}`),
	);
}
