import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import matplotlib.dates as mdates
from .shared import wassersteinDistance

def simpleBenchmarkChart(df: pd.DataFrame, httpReqSending: pd.DataFrame):
  fig, axes = plt.subplots(figsize=(15, 8), nrows=2, ncols=1)
  fig.subplots_adjust(hspace=0.6)
  startTime = df["time"].min()
  endTime = df["time"].max()
  max_response_time = df["value"].max()

  responseTimesAx = axes[0]
  responseTimesAx.set_xlabel("Time")
  responseTimesAx.set_ylabel("Response Time (ms)")
  responseTimesAx.set_ylim(0, min(max_response_time, 1200))
  responseTimesAx.set_xlim(startTime, endTime)

  successDf = df.query("is_error == False").copy()

  sns.scatterplot(data=successDf, x="time", y="value", ax=responseTimesAx, s=2, hue="port", palette="tab20", legend=False,edgecolor="face")
  errorAx = responseTimesAx.twinx()
  errorRate = df.resample("1s", on="time")["is_error"].apply(lambda x: x[x == True].count() / x.count() if x.count() > 0 else 0).reset_index()
  errorAx.set_ylabel("Error Rate")
  errorAx.set_ylim(-0.01, 1)
  errorAx.set_xlim(startTime, endTime)
  errorAx.plot(errorRate["time"], errorRate["is_error"], color='tab:red', linestyle='--')
  errorAx.tick_params(axis='y', colors='red')

  axes[1].set_title(f"rps")
  rps = httpReqSending.resample("1s", on="time")["value"].count().reset_index()
  rps.plot(x="time", y="value", ax=axes[1], legend=False)
  axes[1].set_xlim(startTime, endTime)
  return fig, axes

def responseTimeDistributionScatterPlot(df: pd.DataFrame, ax, startTime, endTime,color="blue"):
  successDf = df.query("is_error == False")
  sns.scatterplot(data=successDf, x="time", y="value", ax=ax, s=2, color=color, edgecolor="face")
  errorRate = df.resample("1s", on="time")["is_error"].apply(lambda x: x[x == True].count() / x.count() if x.count() > 0 else 0).reset_index()
  errorAx = ax.twinx()
  errorAx.set_ylabel("Error Rate")
  errorAx.set_ylim(-0.01, 1)
  errorAx.set_xlim(startTime, endTime)
  errorAx.plot(errorRate["time"], errorRate["is_error"], color='tab:red', linestyle='--')
  errorAx.tick_params(axis='y', colors='red')
  ax.xaxis.set_major_formatter(mdates.DateFormatter('%H:%M'))
  ax.set_yscale('log')
  return ax

def responseTimeDistributionLinePlot(df: pd.DataFrame, ax, startTime, endTime,color="blue"):
  successDf = df.query("is_error == False")
  medians = successDf.resample("5s", on="time")[["value"]].quantile(0.5).reset_index()
  p95 = successDf.resample("5s", on="time")[["value"]].quantile(0.95).reset_index()
  p99 = successDf.resample("5s", on="time")[["value"]].quantile(0.99).reset_index()
  medians["p95"] = p95["value"]
  medians["p99"] = p99["value"]
  sns.lineplot(data=medians, x="time", y="value", ax=ax, color=color)
  if len(medians) > 0:
    ax.fill_between(medians["time"], medians["value"], medians["p95"], alpha=0.4, color=color)
    ax.fill_between(medians["time"], medians["p95"], medians["p99"], alpha=0.2, color=color)
  errorRate = df.resample("1s", on="time")["is_error"].apply(lambda x: x[x == True].count() / x.count() if x.count() > 0 else 0).reset_index()
  errorAx = ax.twinx()
  errorAx.set_ylabel("Error Rate")
  errorAx.set_ylim(-0.01, 1)
  errorAx.set_xlim(startTime, endTime)
  errorAx.plot(errorRate["time"], errorRate["is_error"], color='tab:red', linestyle='--')
  errorAx.tick_params(axis='y', colors='red')
  ax.xaxis.set_major_formatter(mdates.DateFormatter('%H:%M'))
  ax.set_yscale('log')
  return ax

def validationComparisonPlot(httpRequestDuration: pd.DataFrame, mockHttpRequestDuration: pd.DataFrame, httpReqSending: pd.DataFrame):
  combined = pd.concat([httpRequestDuration, mockHttpRequestDuration])
  errorRate = combined.resample("5s", on="time")
  errorRate = errorRate.apply(wassersteinDistance).dropna()
  print(f"total waterstein distance: {errorRate.sum()}")

  fig, axes = plt.subplots(figsize=(15, 8), nrows=3, ncols=1, sharex=True)
  # fig.subplots_adjust(hspace=0.6)

  startTime =  httpRequestDuration["time"].min()
  endTime = httpRequestDuration["time"].max()

  realResponseTimeAx = axes[0]
  responseTimeDistributionScatterPlot(httpRequestDuration, realResponseTimeAx, startTime, endTime)
  # sns.scatterplot(data=httpRequestDuration, x="time", y="value", ax=realResponseTimeAx, color="blue", s=10)
  realResponseTimeAx.set_ylabel("Response Time (ms)")
  realResponseTimeAx.set_xlim(startTime, endTime)

  rpsAx = axes[1]
  rpsAx.set_ylabel("RPS")
  rpsAx.set_xlim(startTime, endTime)
  rps = httpReqSending.resample("1s", on="time")["value"].count()
  sns.lineplot(data=rps, ax=rpsAx, color="green")
  rpsAx.xaxis.set_major_formatter(mdates.DateFormatter('%H:%M'))

  errorRateAx = rpsAx.twinx()
  errorRateAx.set_ylabel("Wasserstein Distance")
  sns.lineplot(data=errorRate, ax=errorRateAx, color="red")
  errorRateAx.fill_between(errorRate.index, errorRate, alpha=0.4, color="red")
  errorRateAx.set_xlim(startTime, endTime)
  errorRateAx.tick_params(axis='y', colors='red')
  # errorRateAx.set_yscale('log')

  mockResponseTimeAx = axes[2]
  responseTimeDistributionScatterPlot(mockHttpRequestDuration, mockResponseTimeAx, startTime, endTime, color="orange")
  mockResponseTimeAx.set_xlabel("Time")
  mockResponseTimeAx.set_ylabel("Response Time (ms)")
  mockResponseTimeAx.set_xlim(startTime, endTime)

  max_y_lim = max(realResponseTimeAx.get_ylim()[1], mockResponseTimeAx.get_ylim()[1])
  realResponseTimeAx.set_ylim(1, max_y_lim)
  mockResponseTimeAx.set_ylim(1, max_y_lim)

  ## add container infos
  startTime = httpRequestDuration["time"].min()
  totalTime = httpRequestDuration["time"].max() - httpRequestDuration["time"].min()
  oneFraction = totalTime / 25

  stages = {
    "one": {
      "end": startTime + 8 * oneFraction,
      "mid": startTime + 4 * oneFraction,
      "start": startTime,
      "containers": 1,
    },
    "two": {
      "end": startTime + 14 * oneFraction,
      "mid": startTime + 11 * oneFraction,
      "start": startTime + 8 * oneFraction,
      "containers": 2
    },
    "three": {
      "end": httpRequestDuration["time"].max(),
      "mid": startTime + 16.5 * oneFraction,
      "start": startTime + 14 * oneFraction,
      "containers": 5
    }
  }
  for stage in stages.values():
    rpsAx.text(stage["start"] + 0.25*oneFraction, rpsAx.get_ylim()[1] * 0.85, f'{stage["containers"]} {"replicas" if stage["containers"] > 1 else "replica"}', color='k')
    rpsAx.axvline(stage["end"], color='b', linestyle='--')
  return fig, axes