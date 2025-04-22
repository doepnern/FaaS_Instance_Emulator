
import common.analysis as analysis
import pandas as pd
import seaborn as sns
import os
import common.lazy as lazy
import scipy.stats as stats
sns.set_theme()

def readAllReplicaResults(runName: str):
  runNumbers = ["1", "32"]
  runs = [analysis.withLoadGrouping(lazy.getHttRequestStats(runName, r)) for r in runNumbers]
  return pd.concat(runs)

def wassersteinDistance(x):
  real = x.query("source == 'real'")
  mock = x.query("source == 'mock'")
  if len(real) == 0 or len(mock) == 0:
    return None
  return stats.wasserstein_distance(real["value"], mock["value"])

for func in ['go-factorial', 'go-io', 'go-string-reversal']:
  realBreakpointRun1 = f"{func}-breakpoint-1"
  realBreakpointRun2 = f"{func}-breakpoint-2"
  mockBreakpointRun1 = f"{func}-breakpoint-mock-1"
  outputFolder = f"output/{func}"
  if not os.path.exists(outputFolder):
    os.makedirs(outputFolder)

  httpsStatsReal1 = readAllReplicaResults(realBreakpointRun1)
  httpsStatsReal2 = readAllReplicaResults(realBreakpointRun2)
  httpStatsMock = readAllReplicaResults(mockBreakpointRun1)

  httpsStatsReal1["source"] = "real"
  httpsStatsReal2["source"] = "mock"
  bothReal = pd.concat([httpsStatsReal1, httpsStatsReal2])
  wasserstein = bothReal.groupby(["load_type", "replicas"])[["load_type", "replicas", "source", "value"]].apply(wassersteinDistance).reset_index(name="wasserstein")
  wasserstein["source"] = "real"
  wasserstein.set_index(["load_type", "replicas","source"], inplace=True)

  httpStatsMock["source"] = "mock"
  realAndMock = pd.concat([httpsStatsReal1, httpStatsMock])
  wassersteinRealAndMock = realAndMock.groupby(["load_type", "replicas"])[["load_type", "replicas", "source", "value"]].apply(wassersteinDistance).reset_index(name="wasserstein")
  wassersteinRealAndMock["source"] = "mock"
  wassersteinRealAndMock.set_index(["load_type", "replicas", "source"], inplace=True)
  print(wassersteinRealAndMock)
  joined = pd.concat([wasserstein, wassersteinRealAndMock])
  joined.reset_index(inplace=True)

  facetGrid = sns.catplot(data=joined, x="replicas", y="wasserstein", hue="source", col="load_type", kind="bar", sharey=False, col_wrap=2)
  facetGrid.savefig(f"{outputFolder}/benchmark_comparison_wasserstein.png")
  facetGrid.savefig(f"{outputFolder}/benchmark_comparison_wasserstein.pdf")



