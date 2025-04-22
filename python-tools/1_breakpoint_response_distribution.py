
import common.utils as utils
import common.analysis as analysis
import pandas as pd
import seaborn as sns
import os
import numpy as np
import common.lazy as lazy
import matplotlib.pyplot as plt
from scipy import stats
import json
sns.set_theme()

for function in ["go-factorial", "go-string-reversal", "go-io"]:
  functionName = f"{function}-breakpoint"
  definition = utils.getBenchmarkDefinition(functionName)
  outputFolder = f"output/{definition['name']}"
  if not os.path.exists(outputFolder):
      os.makedirs(outputFolder)

  httpsStats1 = [analysis.withLoadGrouping(lazy.getHttRequestStats(f"{definition['name']}-breakpoint-{index}", "1")) for index in range(1,4)]
  for ind, df in enumerate(httpsStats1):
      df["run"] = ind
  httpStats1 = pd.concat(httpsStats1)
  summary_table1 = analysis.latencyDistributionByThroughputGroup(httpStats1)
  summary_table1["replicas"] = 1

  httpsStats8 = [analysis.withLoadGrouping(lazy.getHttRequestStats(f"{definition['name']}-breakpoint-{index}", "8")) for index in range(1,4)]
  for ind, df in enumerate(httpsStats8):
      df["run"] = ind
  httpStats8 = pd.concat(httpsStats8)
  summary_table8 = analysis.latencyDistributionByThroughputGroup(httpStats8)
  summary_table8["replicas"] = 8

  httpsStats16 = [analysis.withLoadGrouping(lazy.getHttRequestStats(f"{definition['name']}-breakpoint-{index}", "16")) for index in range(1,4)]
  for ind, df in enumerate(httpsStats16):
      df["run"] = ind
  httpStats16 = pd.concat(httpsStats16)
  summary_table16 = analysis.latencyDistributionByThroughputGroup(httpStats16)
  summary_table16["replicas"] = 16

  httpsStats32 = [analysis.withLoadGrouping(lazy.getHttRequestStats(f"{definition['name']}-breakpoint-{index}", "32")) for index in range(1,4)]
  for ind, df in enumerate(httpsStats32):
      df["run"] = ind
  httpStats32 = pd.concat(httpsStats32)
  summary_table32 = analysis.latencyDistributionByThroughputGroup(httpStats32)
  summary_table32["replicas"] = 32

  # create the response time boxplot
  allStats = pd.concat([httpStats1, httpStats8, httpStats16, httpStats32]).query("is_error == False")
  boxPlot = sns.boxplot(x="load_type", y="value", hue="replicas", data=allStats)
  boxPlot.set_yscale("log")
  boxPlot.set(xlabel="Throughput level", ylabel="Response time (ms)")
  plt.savefig(f"{outputFolder}/profiling_bench_response_times_{definition['docker_image_name']}.png")
  plt.savefig(f"{outputFolder}/profiling_bench_response_times_{definition['docker_image_name']}.pdf")
  plt.close()

  # create the summary table
  combinedTable = pd.concat([summary_table1, summary_table8, summary_table16, summary_table32])
  combinedTable.set_index(["throughput", "replicas", "rps"], inplace=True)
  combinedTable.sort_index(inplace=True)
  styled = combinedTable.style.format(precision=2).hide(["min_load", "max_load"], axis=1)
  styled.to_latex(f"{outputFolder}/summary_table_{definition['docker_image_name']}_multi.tex",  caption=f"Summary table for response time metrics for \\textit{{{definition['docker_image_name']}}}", label=f"tab:summary_table_{definition['docker_image_name']}", position="h", position_float="centering", hrules=True)

  # calculate the response time distribution
  responseDistributions = [
    {
      "s": s,
      "loc": loc,
      "scale": scale,
    } for s, loc, scale in [stats.lognorm.fit(allStats.query(f"load_type == '{load_type}'")["value"]) for load_type in ["minimum", "medium", "large"]]
  ]
  output_file_path = os.path.join(outputFolder, 'function_profile.json')
  print(f"Writing function profile to {output_file_path}")
  with open(output_file_path, 'w') as f:
    json.dump({
      "functionName": functionName,
      "responseDistributions": responseDistributions,
      "maxRps": round(allStats.query("replicas == 1").groupby("load_type", observed=False)["load_max"].mean().loc["large"]),
      "maxHostRps": round(allStats.query("replicas == 8").groupby("load_type", observed=False)["load_max"].mean().loc["large"]),
    }, f)
