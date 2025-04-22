import { DOCKER_SOCKET_PATH } from "../utils";

export async function getDockerInfo() {
	console.log(" fetching docker");
	
	const infoRes = await fetch("http://localhost/info", {
		//@ts-expect-error
		unix: DOCKER_SOCKET_PATH,
	});
	if(!infoRes.ok) throw new Error("Docker info unavailable")

	const info = await infoRes.json();
	console.log("got info");


	const versionRes = await fetch("http://localhost/version", {
		//@ts-expect-error
		unix: DOCKER_SOCKET_PATH,
	});
	const version = await versionRes.json();
	return {
		info,
		version,
	};
}
