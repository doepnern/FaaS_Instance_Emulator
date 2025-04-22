import common.utils as utils
import common.analysis as analysis
import common.charts.breakpoint as charts

import common.lazy as lazy
import matplotlib.pyplot as plt
import os
import seaborn as sns
import pandas as pd
sns.set_theme()

for function in ["go-factorial", "go-string-reversal", "go-io"]:
    functionName = f"{function}-breakpoint"
    definition = utils.getBenchmarkDefinition(functionName)

    for run in range(1, 4):
        definition["name"] = f"{functionName}-{run}"
        print(f"Analyzing {definition['name']}")
        outputFolder = f"output/{definition['name']}"

        if not os.path.exists(outputFolder):
            os.makedirs(outputFolder)
        print(f"Loading request metrics for {definition['name']}")
        runs = [lazy.getResultSummary(definition["name"],r["name"]) for r in definition["benchmarks"]]

        for run in runs:
            print(f"Creating charts for {run['replicas']}")
            httpReqSending = lazy.getHttpReqSending(definition["name"], run["replicas"])
            httpStats = lazy.getHttRequestStats(definition["name"], run["replicas"])
            httpStats = analysis.withLoadGrouping(httpStats)
            print(httpStats.head())
            resourceStats = lazy.getResourceMetrics(definition["name"], run["replicas"])
            fig, axes =charts.responseTimeChart(httpStats, resourceStats, httpReqSending)
            plt.savefig(f"{outputFolder}/response-time-chart_{definition['docker_image_name']}-{run['replicas']}.png", bbox_inches='tight')
            plt.savefig(f"{outputFolder}/response-time-chart_{definition['docker_image_name']}-{run['replicas']}.pdf", bbox_inches='tight')
            plt.close()

    functionName = f"{function}-breakpoint-mock"
    for run in range(1, 2):
        definition["name"] = f"{functionName}-{run}"
        print(f"Analyzing {definition['name']}")
        outputFolder = f"output/{definition['name']}"

        if not os.path.exists(outputFolder):
            os.makedirs(outputFolder)
        print(f"Loading request metrics for {definition['name']}")
        runs = [lazy.getResultSummary(definition["name"],r["name"]) for r in definition["benchmarks"]]

        for run in runs:
            print(f"Creating charts for {run['replicas']}")
            httpReqSending = lazy.getHttpReqSending(definition["name"], run["replicas"])
            httpStats = lazy.getHttRequestStats(definition["name"], run["replicas"])
            httpStats = analysis.withLoadGrouping(httpStats)
            print(httpStats.head())
            resourceStats = lazy.getResourceMetrics(definition["name"], run["replicas"])
            fig, axes = plt.subplots(figsize=(15, 8), nrows=2)
            fig.subplots_adjust(hspace=0.4, wspace=0.3)
            startTime = httpStats["time"].min()
            endTime = httpStats["time"].max()
            charts.responseTimeWithLoadTypesScatterPlot(httpStats, axes[0], startTime, endTime)
            charts.rpsLinePlot(httpStats, axes[1], startTime, endTime)
            plt.savefig(f"{outputFolder}/response-time-chart_{definition['docker_image_name']}-{run['replicas']}.png", bbox_inches='tight')
            plt.savefig(f"{outputFolder}/response-time-chart_{definition['docker_image_name']}-{run['replicas']}.pdf", bbox_inches='tight')
            plt.close()


