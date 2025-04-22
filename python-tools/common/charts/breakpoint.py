import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import matplotlib.dates as mdates


def responseTimeWithLoadTypesScatterPlot(df: pd.DataFrame, ax, startTime, endTime):
  successDf = df.query("is_error == False")

  sns.scatterplot(data=successDf, x="time", y="value", ax=ax, s=2, hue="port", palette="tab20", legend=False, edgecolor="face")

  maximumStart = df[df["load_type"] == "maximum"]["time"].min()
  largeStart = df[df["load_type"] == "large"]["time"].min()
  mediumStart = df[df["load_type"] == "medium"]["time"].min()

  # Plot vertical lines at 25%, 50%, 75% and 100% of the time passed
  ax.axvline(mediumStart, color='b', linestyle='--')
  ax.axvline(largeStart, color='b', linestyle='--')
  ax.axvline(maximumStart, color='b', linestyle='--')
  ax.axvline(endTime, color='b', linestyle='--')

  # fin middle timestamps between lines
  mid_min = startTime + (mediumStart - startTime) / 2
  mid_min_med = mediumStart + (largeStart - mediumStart) / 2
  mid_med_large = largeStart + (maximumStart - largeStart) / 2

  # Adding text annotations for each load boundary
  ax.text(mid_min, ax.get_ylim()[1] * 0.85, 'Minimum Load', color='k', ha='center', fontsize=10)
  ax.text(mid_min_med, ax.get_ylim()[1] * 0.85, 'Medium Load', color='k', ha='center', fontsize=10)
  ax.text(mid_med_large, ax.get_ylim()[1] * 0.85, 'Large Load', color='k', ha='center', fontsize=10)

  ax.set_xlabel("Time")
  ax.set_ylabel("Response Time (ms)")

  max_response_time = df["value"].max()
  ax.set_ylim(0, min(max_response_time, 1000))
  ax.set_xlim(startTime, endTime)
  ax.xaxis.set_major_formatter(mdates.DateFormatter('%H:%M'))
  return ax


def rpsLinePlot(httpReqSending: pd.DataFrame, ax, startTime, endTime):
  ax.set_title(f"rps")
  rps = httpReqSending.resample("1s", on="time")["value"].count().reset_index()
  rps.plot(x="time", y="value", ax=ax, legend=False)
  ax.set_xlim(startTime, endTime)
  ax.set_xlabel("Time")
  ax.set_ylabel("rps")
  return ax

def responseErrorRateLinePlot(df: pd.DataFrame, ax, startTime, endTime):
  errorRate = df.resample("1s", on="time")["is_error"].apply(lambda x: x[x == True].count() / x.count() if x.count() > 0 else 0).reset_index()
  ax.set_ylabel("Error Rate")
  ax.set_ylim(-0.01, 1)
  ax.set_xlim(startTime, endTime)
  ax.plot(errorRate["time"], errorRate["is_error"], color='tab:red', linestyle='--')
  ax.tick_params(axis='y', colors='red')
  return ax



def responseTimeChart(df: pd.DataFrame, resource_metrics: pd.DataFrame , httpReqSending: pd.DataFrame):
  fig, axes = plt.subplots(figsize=(15, 8), nrows=2, ncols=2)
  fig.subplots_adjust(hspace=0.4, wspace=0.3)
  startTime = df["time"].min()
  endTime = df["time"].max()

  ax = axes[0][0]
  responseTimeWithLoadTypesScatterPlot(df, ax, startTime, endTime)

  ax2 = ax.twinx()
  responseErrorRateLinePlot(df, ax2, startTime, endTime)

  ax3 = axes[1][0]
  rpsLinePlot(httpReqSending, ax3, startTime, endTime)

  axes[0][1].set_title(f"CPU usage")
  hostCpuLimit = 400
  containerCpuLimit = 100
  cpuUsagePerContainer = resource_metrics.query('metric_name == "cpuPerContainer"').copy()
  sns.lineplot(data=cpuUsagePerContainer, x="time", y="value", ax=axes[0][1], label="Per Container CPU Usage", color='g')
  cpuContainerLimitDf = pd.DataFrame(data={"time": [startTime, endTime], "value": [containerCpuLimit, containerCpuLimit]})
  sns.lineplot(data=cpuContainerLimitDf, x="time", y="value", ax=axes[0][1], label="Container CPU Limit", color='g', linestyle='--')
  totalCpuUsage = cpuUsagePerContainer.resample("5s", on="time")["value"].sum().reset_index()
  sns.lineplot(data=totalCpuUsage, x="time", y="value", ax=axes[0][1], label="Total CPU Usage", color='b')
  cpuHostLimitDf = pd.DataFrame(data={"time": [startTime, endTime], "value": [hostCpuLimit, hostCpuLimit]})
  sns.lineplot(data=cpuHostLimitDf, x="time", y="value", ax=axes[0][1], label="Host CPU Limit", color='b', linestyle='--')
  axes[0][1].set_xlim(startTime, endTime)
  axes[0][1].set_xlabel("Time")
  axes[0][1].set_ylabel("CPU usage (%)")
  axes[0][1].xaxis.set_major_formatter(mdates.DateFormatter('%H:%M'))

  axes[1][1].set_title(f"Memory usage")
  containerMemoryLimitInMb = 512
  hostMemoryLimitInMb = 16 * 1024  # 16GB
  # memory usage in bytes
  memoryUsagePerContainer = resource_metrics.query('metric_name == "memoryPerContainer"').copy()
  # transform to MB
  memoryUsagePerContainer["value"] = memoryUsagePerContainer["value"] / 1024 / 1024
  totalMemoryUsage = memoryUsagePerContainer.resample("5s", on="time")["value"].sum().reset_index()
  sns.lineplot(data=memoryUsagePerContainer, x="time", y="value", ax=axes[1][1], label="Per Container Memory Usage", color='g')
  sns.lineplot(data=totalMemoryUsage, x="time", y="value", ax=axes[1][1], label="Total Memory Usage", color='b')
  containerMemoryLimitDf = pd.DataFrame(data={"time": [startTime, endTime], "value": [containerMemoryLimitInMb, containerMemoryLimitInMb]})
  sns.lineplot(data=containerMemoryLimitDf, x="time", y="value", ax=axes[1][1], label="Container Memory Limit", color='g', linestyle='--')
  axes[1][1].set_xlim(startTime, endTime)
  axes[1][1].set_xlabel("Time")
  axes[1][1].set_ylabel("Memory usage (MB)")
  axes[1][1].xaxis.set_major_formatter(mdates.DateFormatter('%H:%M'))
  return fig, axes