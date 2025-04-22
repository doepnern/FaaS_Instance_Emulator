import { writeFileSync } from "node:fs";
import { add } from "date-fns";
const PROMETHEUS_URL =
	import.meta.env.PROMETHEUS_URL || "http://localhost:9090";

const cpuUsagePerCpu = {
	query: `(sum by (cpu)(rate(node_cpu_seconds_total{mode!="idle"}[5s]))*100)`,
	name: "cpuPerCore",
};
const cpuUsageTotal = {
	query: `(sum(rate(node_cpu_seconds_total{mode!="idle"}[5s]))*100)`,
	name: "cpuTotal",
};

const cpuUsagePerContainer = {
	query: `sum(rate(container_cpu_usage_seconds_total{name=~".+"}[5s])) by (name) * 100`,
	name: "cpuPerContainer",
};

const totalCpuUsageContainers = {
	query: `sum(rate(container_cpu_system_seconds_total{name=~".+"}[5s]))`,
	name: "totalCpuUsageContainers",
};

const memoryUsagePerContainer = {
	query: `sum by (name)(rate(container_memory_rss{name=~".+"}[5s]))`,
	name: "memoryPerContainer",
}
const totalMemoryUsageContainers = {
	query: `sum(rate(container_memory_rss{name=~".+"}[5s]))`,
	name: "totalMemoryUsageContainers",
}

const sentNetworkPerContainer = {
	query: `sum(rate(container_network_transmit_bytes_total{name=~".+"}[5s])) by (name)`,
	name: "sentNetworkPerContainer",
}

const receivedNetworkPerContainer = {
	query: `sum(rate(container_network_receive_bytes_total{name=~".+"}[5s])) by (name)`,
	name: "receivedNetworkPerContainer",
}

const memoryUsagePercent = {
	query:
		"((node_memory_MemTotal_bytes-node_memory_MemAvailable_bytes)/node_memory_MemTotal_bytes*100)",
	name: "memoryUsagePercent",
};

const memoryUsage = {
	query: "(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes)",
	name: "memoryUsage",
};

const readKbPerSecond = {
	query: "rate(node_disk_read_bytes_total[5s])",
	name: "readKbPerSecond",
};

const writeKbPerSecond = {
	query: "rate(node_disk_written_bytes_total[5s])",
	name: "writeKbPerSecond",
};

const networkInKbPerSecond = {
	query: "rate(node_network_receive_bytes_total[5s])",
	name: "networkInKbPerSecond",
};

const networkOutKbPerSecond = {
	query: "rate(node_network_transmit_bytes_total[5s])",
	name: "networkOutKbPerSecond",
};

async function fetchQuery(
	metricDef: { query: string; name: string },
	{ from, until }: { from: Date; until: Date },
) {
	const encoded = encodeURIComponent(metricDef.query);
	const res = await fetch(
		`${PROMETHEUS_URL}/api/v1/query_range?query=${encoded}&start=${from.toISOString()}&end=${until.toISOString()}&step=5s`,
	);
	if (!res.ok) {
		console.error(res);
		const jsonBody = await res.json();
		console.error(jsonBody);
		throw new Error("Failed to fetch data from prometheus");
	}
	const data = (await res.json()) as {
		data: {
			result: { values: [number, string][]; metric: Record<string, any> }[];
		};
	};
	if (metricDef.name === "cpuPerContainer") {
		writeFileSync("cpuPerContainer.json", JSON.stringify(data, null, 2));
	}
	const returnData = data.data.result.map((r) => ({
		...r,
		values: r.values.map(([time, value]) => ({
			time: new Date(time * 1000),
			value,
		})),
		metric: { ...r.metric, name: metricDef.name, id: r.metric.name },
	}));
	return returnData;
}

const queries = [
	cpuUsagePerCpu,
	cpuUsageTotal,
	memoryUsagePercent,
	memoryUsage,
	readKbPerSecond,
	writeKbPerSecond,
	networkInKbPerSecond,
	networkOutKbPerSecond,
	cpuUsagePerContainer,
	totalCpuUsageContainers,
	memoryUsagePerContainer,
	totalMemoryUsageContainers,
	sentNetworkPerContainer,
	receivedNetworkPerContainer
]

export async function collectMetrics({
	from,
	until,
}: {
	from: Date;
	until: Date;
}) {
	const metrics = await Promise.all(queries.map((query) => fetchQuery(query, { from, until })))
	return metrics.flat();
}

const start = add(new Date(), { hours: -1 });
const end = new Date();

const metrics = await collectMetrics({
	from: start,
	until: end,
});

writeFileSync("metrics.json", JSON.stringify(metrics, null, 2));
