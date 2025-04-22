export function fetchDockerApi(path: string, init?: RequestInit) {
	const DOCKER_HOST = import.meta.env.DOCKER_HOST as string;
	if (!DOCKER_HOST) {
		throw new Error("DOCKER_HOST is not defined");
	}
	const isUnix = DOCKER_HOST.startsWith("unix://");
	return isUnix
		? fetch(`http://localhost:2375${path}`, {
			// @ts-ignore
			unix: "/var/run/docker.sock",
			...init,
		})
		: fetch(`http://${DOCKER_HOST}/${path}`, init);
}
