from typing import List, TypedDict
import json
import pandas as pd
import os
from . import analysis

class Benchmark(TypedDict):
    rps: int

class FunctionBenchmarkConfig(TypedDict):
    name: str
    url: str
    docker_container_name: str
    benchmarks: List[Benchmark]

class FunctionBenchmarkResult(TypedDict):
    started_at: str
    ended_at: str
    k6MetricsFile: str
    prometheusMetricsFile: str
    dockerMetricsFile: str
    summary: dict

class PrometheusMetricInfo(TypedDict):
    name: str

class PrometheusMetric(TypedDict):
    metric: PrometheusMetricInfo
    values: List[dict]


def metricToDf(metric: PrometheusMetric) -> pd.DataFrame:
    values = [{**x, "metric_name": metric["metric"]["name"], "metric": metric["metric"]} for x in metric["values"]]
    df = pd.DataFrame(values)
    df["time"] = pd.to_datetime(df["time"], utc=True)
    df["value"] = pd.to_numeric(df["value"], errors='raise')
    return df\


class FunctionBenchmarkResultsType(TypedDict):
    request_metrics: pd.DataFrame
    resource_metrics: pd.DataFrame
    resource_metrics_docker: pd.DataFrame
    result: FunctionBenchmarkResult

def getFunctionBenchmarkResults(functionBenchmarkConfig: FunctionBenchmarkConfig):
    resultDict :  FunctionBenchmarkResultsType = {}
    for benchmark in functionBenchmarkConfig["benchmarks"]:
        path = os.path.realpath(f'../benchmarks/results/{functionBenchmarkConfig["name"]}/{benchmark["name"]}_result.json')
        if not os.path.exists(path):
            print(f"Warning: Missing result file for {benchmark['name']} at path {path}")
            continue
        with open(f'../benchmarks/results/{functionBenchmarkConfig["name"]}/{benchmark["name"]}_result.json', 'r') as file:
            result = json.load(file)
            name = benchmark["name"]
            replicas = benchmark["replicas"] if "replicas" in benchmark else 1
            if not os.path.exists(result["k6MetricsFile"]):
                print(f"Warning: Missing k6 metrics files for {name}")
                continue
            # if not os.path.exists(result["dockerMetricsFile"]):
            #     print(f"Warning: Missing docker metrics files for {name}")
            #     continue
            if not os.path.exists(result["prometheusMetricsFile"]):
                print(f"Warning: Missing prometheus metrics files for {name}")
                continue
            k6Metrics = pd.read_json(result["k6MetricsFile"], lines=True)
            # dockerMetrics = pd.read_json(result["dockerMetricsFile"], lines=True)
            prometheusData : List[PrometheusMetric] = []
            with open(result["prometheusMetricsFile"], 'r') as prometheusFile:
                prometheusData = json.load(prometheusFile)
            prometheusMetrics = pd.concat([metricToDf(x) for x in prometheusData])
            k6Metrics["testName"] = name
            k6Metrics["replicas"] = int(replicas)
            k6Metrics.sort_values(by=["replicas"], inplace=True)
            # dockerMetrics["testName"] = name
            # dockerMetrics["replicas"] = int(replicas)
            # dockerMetrics["read"] = pd.to_datetime(dockerMetrics["read"], utc=True)
            # dockerMetrics.sort_values(by=["replicas", "read"], inplace=True)
            prometheusMetrics["testName"] = name
            prometheusMetrics["replicas"] = int(replicas)
            prometheusMetrics.sort_values(by=["replicas", "time"], inplace=True)
            resultDict[name] = {}
            resultDict[name]["request_metrics"] = k6Metrics
            # resultDict[name]["resource_metrics_docker"] = dockerMetrics
            resultDict[name]["resource_metrics"] = prometheusMetrics
            resultDict[name]["result"] = result
    return resultDict

def getBenchmarkDefinition(functionName: str) -> List[FunctionBenchmarkConfig]:
    with open(f'../benchmarks/targets/{functionName}.json', 'r') as file:
        definition : FunctionBenchmarkConfig = json.load(file)
    return definition


def getBreaktestResults(definition: FunctionBenchmarkConfig) -> pd.DataFrame:
        dockerDf = pd.DataFrame(
        columns=["testName", "replicas"]
        )
        k6Df = pd.DataFrame(
            columns=["testName", "replicas"]
        )
        for benchmark in definition["benchmarks"]:
            name = benchmark["name"]
            replicas = benchmark["replicas"] if "replicas" in benchmark else 1
            k6MetricsPath = f'../benchmarks/results/{definition["name"]}/{name}_metrics'
            dockerMetricsPath = f'../benchmarks/results/{definition["name"]}/{name}_docker-metrics'
            if not os.path.exists(k6MetricsPath) or not os.path.exists(dockerMetricsPath):
                print(f"Warning: Missing metrics files for {name}")
                continue
            k6Metrics = pd.read_json(k6MetricsPath, lines=True)
            dockerMetrics = pd.read_json(dockerMetricsPath, lines=True)
            k6Metrics["testName"] = name
            k6Metrics["replicas"] = int(replicas)
            dockerMetrics["testName"] = name
            k6Metrics["replicas"] = int(replicas)
            k6Df = pd.concat([k6Df, k6Metrics])
            dockerDf = pd.concat([dockerDf, dockerMetrics])
            dockerDf["read"] = pd.to_datetime(dockerDf["read"], utc=True)
        k6Df.sort_values(by=["replicas"], inplace=True)
        dockerDf.sort_values(by=["replicas", "read"], inplace=True)
        return {
            "k6": k6Df,
            "docker": dockerDf
        }


def getBenchmarkResults(definition: FunctionBenchmarkConfig) -> pd.DataFrame:
    dockerDf = pd.DataFrame(
        columns=["rps"]
    )
    k6Df = pd.DataFrame(
        columns=["rps"]
    )
    for funcDef in definition["benchmarks"]:
        k6Metrics = pd.read_json(f'../benchmarks/results/{definition["name"]}/{funcDef["rps"]}rps_metrics', lines=True)
        dockerMetrics = pd.read_json(f'../benchmarks/results/{definition["name"]}/{funcDef["rps"]}rps_docker-metrics', lines=True)
        k6Metrics["rps"] = funcDef["rps"]
        dockerMetrics["rps"] = funcDef["rps"]
        k6Df = pd.concat([k6Df, k6Metrics])
        dockerDf = pd.concat([dockerDf, dockerMetrics])
        dockerDf["read"] = pd.to_datetime(dockerDf["read"], utc=True)
    return {
        "k6": k6Df,
        "docker": dockerDf
    }


def dataJsonToSeries(s: pd.Series) -> pd.Series:
    df = pd.json_normalize(s["data"])
    df["time"] = pd.to_datetime(df["time"])
    if(len(df) == 0):
        raise ValueError("No data in the series")
    return df.iloc[0]

def normalizeColumn(series: pd.Series, column: str) -> pd.Series:
    data = series[column]
    if isinstance(data, dict):
        #combine the series with the original series
        combined = pd.Series({**series, **data})
        return combined

def getHttRequestStats(k6Df: pd.DataFrame) -> pd.DataFrame:
    http_req_duration = k6Df.query("type == 'Point' and metric == 'http_req_duration'")
    data_df : pd.DataFrame = pd.json_normalize(http_req_duration['data'])
    data_df["time"] = pd.to_datetime(data_df["time"], utc=True)
    data_df["testName"] = http_req_duration["testName"].values
    data_df["replicas"] = http_req_duration["replicas"].values
    data_df["port"] = data_df["tags.name"].apply(lambda x: x.split(":")[-1])
    data_df["is_error"] = data_df.apply(lambda x: x["value"] > 1000 or x["tags.status"] != "200", axis=1)
    data_df["current_rps"] = data_df["tags.currentRps"].apply(lambda x: float(x))
    data_df.sort_values(by=["replicas","port","time"], inplace=True)
    return data_df

def getHttpReqSending(k6Df: pd.DataFrame) -> pd.DataFrame:
    http_req_sending = k6Df.query("type == 'Point' and metric == 'http_req_sending'")
    data_df : pd.DataFrame = pd.json_normalize(http_req_sending['data'])
    data_df["time"] = pd.to_datetime(data_df["time"], utc=True)
    data_df["testName"] = http_req_sending["testName"].values
    data_df["replicas"] = http_req_sending["replicas"].values
    data_df["port"] = data_df["tags.name"].apply(lambda x: x.split(":")[-1])
    data_df.sort_values(by=["replicas","port","time"], inplace=True)
    return data_df

def getCpuStats(dockerDf: pd.DataFrame) -> pd.DataFrame:
    return dockerDf.apply(lambda x: normalizeColumn(x, "container_info"), axis=1)




