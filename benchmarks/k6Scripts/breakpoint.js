import http from "k6/http";
import exec from "k6/execution";
import { check } from "k6";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.1/index.js";

/**
 * @type {import("./src/types").MultiRpsEnv}
 */
const {
	BENCHMARK_TEST_NAME = "k6",
	FUNC_NAME,
	PAYLOAD,
	HOST,
	PORT_RANGE,
	PATH = "",
	RESPONSE,
	BENCHMARK
} = __ENV;
const [startPort, endPort] = PORT_RANGE.split("-").map((p) =>
	Number.parseInt(p),
);

const benchmark = JSON.parse(BENCHMARK);
const breakpointStage = { duration: '5m', target: 2000, ...benchmark.options };

export const options = {
	scenarios: {
		breakpoint: {
			executor: 'ramping-arrival-rate', //Assure load increase if the system slows
			stages: [
				// {
				// 	duration: '10s', target: 0
				// },
				breakpointStage,
			],
			preAllocatedVUs: 2000, // to allocate runtime resources
		},
	},
	thresholds: {
		http_req_failed: [{
			threshold: 'rate<0.01',
			abortOnFail: true, // boolean
			delayAbortEval: '0s', // string
		}], // http errors should be less than 1%
		http_req_duration: [{
			threshold: 'p(95)<1000',
			abortOnFail: true, // boolean
			delayAbortEval: '0s', // string
		}], // 95% of requests should be below 200ms
	},
};

/**
 * @param {number} iterationInTest
 */
function getNextPort(iterationInTest) {
	// if(Math.random() > 0.5) return startPort
	return startPort + (iterationInTest % (endPort - startPort + 1));
}

const expectedResponseString = RESPONSE
	? JSON.stringify(JSON.parse(RESPONSE))
	: undefined;

function calcRpsByProgress(progress, stage) {
	return progress * stage.target;
}

// The function that defines VU logic.
// See https://grafana.com/docs/k6/latest/examples/get-started-with-k6/ to learn more
// about authoring k6 scripts.
//
export default function () {
	const port = getNextPort(exec.scenario.iterationInTest);
	const funcUrl = `${HOST}:${port}${PATH}`;
	const tags = {
		currentRps: Math.round(calcRpsByProgress(exec.scenario.progress, breakpointStage)),
	}
	const res = http.post(funcUrl, JSON.stringify(JSON.parse(PAYLOAD)), {
		headers: { "Content-Type": "application/json" },
		tags
	});
	check(res, {
		"is status 200": (r) => r.status === 200,
		...(expectedResponseString
			? {
				"correct response": (r) => {
					const actual = JSON.parse(r.body);
					return expectedResponseString === JSON.stringify(actual);
				},
			}
			: {}),
	});
}

export function handleSummary(data) {
	return {
		[`./results/${FUNC_NAME}/${BENCHMARK_TEST_NAME}_summary.json`]: JSON.stringify(data),
		stdout: textSummary(data, { indent: " ", enableColors: true }),
	};
}
