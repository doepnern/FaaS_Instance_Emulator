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
  PATH = "",
  RESPONSE,
  BENCHMARK,
  IMAGE_NAME,
} = __ENV;


const benchmark = JSON.parse(BENCHMARK);

const totalDurationInMinutes = benchmark.options?.duration ?? "5m";
const duration = Number.parseInt(totalDurationInMinutes.split("m")[0])
if (Number.isNaN(duration)) {
  throw new Error("Invalid duration provided in the benchmark options")
}

const target = benchmark.options?.maxRps;
if (target === undefined) {
  throw new Error("Target is required for the validation benchmark and needs to be the maximum throughput")
}

const hostMaxRps = benchmark.options?.maxHostRps;
if (hostMaxRps === undefined) {
  throw new Error("Host max RPS is required for the validation benchmark")
}

const oneThroughputPercent = target / 100
const oneHostMaxRps = hostMaxRps / 100

const sharedSettings = {
  executor: 'ramping-arrival-rate', //Assure load increase if the system slows
  preAllocatedVUs: 2000, // to allocate runtime resources
}

// allow a total of 25 stages
const oneFraction = duration / 100 * 4

const stageOne = {
  ...sharedSettings,
  stages: [
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(15 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(15 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(50 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(50 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(100 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(50 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(50 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(80 * oneThroughputPercent)
    }
  ],
  gracefulStop: '0s'
}
const stageOneDuration = stageOne.stages.length * oneFraction
const stageTwo = {
  ...sharedSettings,
  stages: [
    {
      duration: `0m`,
      target: Math.ceil(120 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(120 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(120 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(200 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(140 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(140 * oneThroughputPercent)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(160 * oneThroughputPercent)
    },
  ],
  startTime: `${stageOneDuration}m`,
  gracefulStop: '0s'
}

const stageTwoDuration = (stageTwo.stages.length - 1) * oneFraction

const stageThree = {
  ...sharedSettings,
  stages: [
    {
      duration: `0m`,
      target: Math.ceil(40 * oneHostMaxRps)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(40 * oneHostMaxRps)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(100 * oneHostMaxRps)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(60 * oneHostMaxRps)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(60 * oneHostMaxRps)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(120 * oneHostMaxRps)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(50 * oneHostMaxRps)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(60 * oneHostMaxRps)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(130 * oneHostMaxRps)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(50 * oneHostMaxRps)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(50 * oneHostMaxRps)
    },
    {
      duration: `${oneFraction}m`,
      target: Math.ceil(0)
    },
  ],
  startTime: `${stageOneDuration + stageTwoDuration}m`,
  gracefulStop: '0s'
}

// const stageThreeDuration = (stageThree.stages.length - 1) * oneFraction

export const options = {
  scenarios: {
    "validation-1": stageOne,
    "validation-2": stageTwo,
    "validation-5": stageThree
  },
  thresholds: {
  },
};

const expectedResponseString = RESPONSE
  ? JSON.stringify(JSON.parse(RESPONSE))
  : undefined;

/**
 * @param {number} iterationInTest
 */
function getNextParameter(iterationInTest, scenarioName) {
  const numReplicas = Number.parseInt(scenarioName.split("-")[1])
  if (Number.isNaN(numReplicas)) {
    throw new Error("Invalid scenario name")
  }
  return `?name=${IMAGE_NAME}-${iterationInTest % numReplicas}`;
}

// The function that defines VU logic.
// See https://grafana.com/docs/k6/latest/examples/get-started-with-k6/ to learn more
// about authoring k6 scripts.
//
export default function () {
  const param = getNextParameter(exec.scenario.iterationInTest, exec.scenario.name);
  const funcUrl = `${HOST}:8089${PATH}${param}`;


  const res = http.post(funcUrl, JSON.stringify(JSON.parse(PAYLOAD)), {
    headers: { "Content-Type": "application/json" }
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
