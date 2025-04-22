export type MultiRpsEnv = {
	FUNC_NAME: string;
	HOST: string;
	/** e.g. 3000-3020 */
	PORT_RANGE: string;
	PATH?: string;
	RESPONSE?: string;
	PAYLOAD?: string;
	K6_WEB_DASHBOARD: string;
	BENCHMARK_TEST_NAME?: string;
	/** JSON stringified benchmark definition*/
	BENCHMARK?: string;
	REPLICAS: string;
	IMAGE_NAME: string;
};
