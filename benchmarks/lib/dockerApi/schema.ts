import { z } from "zod";

export const DockerStatsSchemaBase = z.object({
	read: z.string(),
	pids_stats: z
		.object({
			current: z.number(),
		})
		.partial(),
	networks: z
		.record(
			z.object({
				rx_bytes: z.number(),
				rx_dropped: z.number(),
				rx_errors: z.number(),
				rx_packets: z.number(),
				tx_bytes: z.number(),
				tx_dropped: z.number(),
				tx_errors: z.number(),
				tx_packets: z.number(),
			}),
		)
		.optional(),
	memory_stats: z
		.object({
			stats: z.object({
				total_pgmajfault: z.number(),
				cache: z.number(),
				mapped_file: z.number(),
				total_inactive_file: z.number(),
				pgpgout: z.number(),
				rss: z.number(),
				total_mapped_file: z.number(),
				writeback: z.number(),
				unevictable: z.number(),
				pgpgin: z.number(),
				total_unevictable: z.number(),
				pgmajfault: z.number(),
				total_rss: z.number(),
				total_rss_huge: z.number(),
				total_writeback: z.number(),
				total_inactive_anon: z.number(),
				rss_huge: z.number(),
				hierarchical_memory_limit: z.number(),
				total_pgfault: z.number(),
				total_active_file: z.number(),
				active_anon: z.number(),
				total_active_anon: z.number(),
				total_pgpgout: z.number(),
				total_cache: z.number(),
				inactive_anon: z.number(),
				active_file: z.number(),
				pgfault: z.number(),
				inactive_file: z.number(),
				total_pgpgin: z.number(),
			}).partial(),
			max_usage: z.number(),
			usage: z.number(),
			failcnt: z.number().optional(),
			limit: z.number(),
		})
		.partial(),
	blkio_stats: z.object({}).passthrough(),
	cpu_stats: z.object({
		cpu_usage: z.object({
			percpu_usage: z.array(z.number()).optional(),
			usage_in_usermode: z.number(),
			total_usage: z.number(),
			usage_in_kernelmode: z.number(),
		}),
		system_cpu_usage: z.number().optional(),
		online_cpus: z.number().optional(),
		throttling_data: z.object({
			periods: z.number(),
			throttled_periods: z.number(),
			throttled_time: z.number(),
		}),
	}),
	precpu_stats: z.object({
		cpu_usage: z
			.object({
				percpu_usage: z.array(z.number()),
				usage_in_usermode: z.number(),
				total_usage: z.number(),
				usage_in_kernelmode: z.number(),
			})
			.partial(),
		system_cpu_usage: z.number().optional(),
		online_cpus: z.number().optional(),
		throttling_data: z.object({
			periods: z.number(),
			throttled_periods: z.number(),
			throttled_time: z.number(),
		}),
	}),
});

function getCpuMetrics(data: z.infer<typeof DockerStatsSchemaBase>) {
	const previousSystemCpu = data.precpu_stats.system_cpu_usage;
	const systemCpu = data.cpu_stats.system_cpu_usage;
	const number_cpus = data.cpu_stats.online_cpus;

	if (
		previousSystemCpu === undefined ||
		systemCpu === undefined ||
		number_cpus === undefined
	)
		return {};
	const cpu_delta =
		data.cpu_stats.cpu_usage.total_usage -
		(data.precpu_stats.cpu_usage.total_usage ?? 0);
	const system_cpu_delta = systemCpu - previousSystemCpu;
	const cpu_usage_percent =
		(cpu_delta / system_cpu_delta) * number_cpus * 100.0;
	return {
		cpu_delta,
		system_cpu_delta,
		cpu_usage_percent,
	};
}

function getMemoryMetrics(data: z.infer<typeof DockerStatsSchemaBase>) {
	const memoryStatsUsage = data.memory_stats.usage;
	// Assume we are in cgroup v2, see https://docs.docker.com/reference/cli/docker/container/stats/
	const memoryStatsCache = data.memory_stats.stats?.inactive_file;
	if (memoryStatsUsage === undefined || memoryStatsCache === undefined || data.memory_stats.limit === undefined)
		return {};
	const used_memory = memoryStatsUsage - memoryStatsCache;
	const available_memory = data.memory_stats.limit;
	return {
		used_memory,
		available_memory,
		memory_usage_percent: (used_memory / available_memory) * 100.0,
	};
}

export const DockerStatsSchema = DockerStatsSchemaBase.transform((data) => {
	return {
		...data,
		...getCpuMetrics(data),
		...getMemoryMetrics(data),
	};
});

export type DockerStatsMetric = z.infer<typeof DockerStatsSchema>;
