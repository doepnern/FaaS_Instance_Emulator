from typing import TypedDict, List
import pandas as pd
import json


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

class ResultSummary(TypedDict):
  startedAt: str
  endedAt: str
  k6MetricsFile: str
  prometheusMetricsFile: int
  dockerMetricsFile: int
  summary: dict
  replicas: str


def getResultSummary(funcName: str, run: str) -> ResultSummary:
  with open(f'../benchmarks/results/{funcName}/{run}_result.json', 'r') as file:
    result = json.load(file)
  result["replicas"] = run
  return result

httpRequestStatsCols = ["time", "value", "testName", "replicas", "port", "is_error", "current_rps"]

def getHttRequestStats(funcName: str, testName: str) -> pd.DataFrame:
  allRows = []
  for chunk in pd.read_json(f'../benchmarks/results/{funcName}/{testName}_metrics.json', lines=True, chunksize=10000):
    relevantRows = chunk.query("type == 'Point' and metric == 'http_req_duration'")
    if len(relevantRows) == 0:
      continue
    allRows.append(relevantRows[['data']])
  data_df = pd.concat(allRows)
  data_df : pd.DataFrame = pd.json_normalize(data_df['data'])
  data_df["time"] = pd.to_datetime(data_df["time"], utc=True)
  data_df["testName"] = testName
  data_df["replicas"] = int(testName)
  data_df["port"] = data_df["tags.name"].apply(lambda x: x.split(":")[-1])
  data_df["is_error"] = data_df.apply(lambda x: x["value"] > 1000 or x["tags.status"] != "200", axis=1)
  if "tags.currentRps" in data_df.columns:
    data_df["current_rps"] = data_df["tags.currentRps"].apply(lambda x: float(x))

  data_df.sort_values(by=["replicas","port","time"], inplace=True)
  return data_df

def getHttpReqSending(funcName: str, testName: str) -> pd.DataFrame:
  allRows = []
  for chunk in pd.read_json(f'../benchmarks/results/{funcName}/{testName}_metrics.json', lines=True, chunksize=10000):
    relevantRows = chunk.query("type == 'Point' and metric == 'http_req_sending'")
    if len(relevantRows) == 0:
      continue
    allRows.append(relevantRows[['data']])
  data_df = pd.concat(allRows)
  data_df : pd.DataFrame = pd.json_normalize(data_df['data'])
  data_df["time"] = pd.to_datetime(data_df["time"], utc=True)
  data_df["testName"] = testName
  data_df["replicas"] = int(testName)
  data_df["port"] = data_df["tags.name"].apply(lambda x: x.split(":")[-1])
  data_df.sort_values(by=["replicas","port","time"], inplace=True)
  return data_df


def getResourceMetrics(funcName: str, testName: str) -> pd.DataFrame:
  with open(f'../benchmarks/results/{funcName}/{testName}_metrics_prometheus.json', 'r') as file:
      prometheusData = json.load(file)
  prometheusMetrics = pd.concat([metricToDf(x) for x in prometheusData])
  prometheusMetrics["testName"] = testName
  prometheusMetrics["replicas"] = int(testName)
  prometheusMetrics["id"] = prometheusMetrics["metric"].apply(lambda x: x["id"] if "id" in x else "default")
  prometheusMetrics.sort_values(by=["replicas", "time"], inplace=True)
  return prometheusMetrics