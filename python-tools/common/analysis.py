import pandas as pd


def defineThroughputGroups(df: pd.DataFrame):
  load = df.query("is_error == False").resample("5s", on="time").count() / 5
  maxLoad= load["value"].max() * 0.95
  maxDesiredLoad = df["current_rps"].max()
  maximumloadMaxValue = maxDesiredLoad
  minimumloadMaxValue = round(maxLoad / 3)
  mediumLoadMaxValue = round((maxLoad / 3)* 2)
  largeLoadMaxValue = round(maxLoad)
  throughputDict = {
    "minimum": {
      "min": 0,
      "max": minimumloadMaxValue,
    },
    "medium": {
      "min": minimumloadMaxValue,
      "max": mediumLoadMaxValue,
    },
    "large": {
      "min": mediumLoadMaxValue,
      "max": largeLoadMaxValue,
    },
    "maximum": {
      "min": largeLoadMaxValue,
      "max": maximumloadMaxValue,
    }
  }
  return throughputDict

def withLoadGrouping(df: pd.DataFrame):
  throughputDict = defineThroughputGroups(df)
  df = df.copy()
  df["load_type"] = df.apply(lambda x: addLoadGrouping(x, throughputDict), axis=1)
  df["load_type"] = pd.Categorical(df["load_type"], categories=["minimum", "medium", "large", "maximum"], ordered=True)
  df["load_min"] = df.apply(lambda x:throughputDict[x["load_type"]]["min"], axis=1)
  df["load_max"] = df.apply(lambda x:throughputDict[x["load_type"]]["max"], axis=1)
  return df

def addLoadGrouping(row: pd.Series, throughputDict: dict):
  if row["current_rps"] <= throughputDict["minimum"]["max"]:
    return "minimum"
  elif row["current_rps"] <= throughputDict["medium"]["max"]:
    return "medium"
  elif row["current_rps"] <= throughputDict["large"]["max"]:
    return "large"
  elif row["current_rps"] > throughputDict["large"]["max"]:
    return "maximum"
  else:
    raise ValueError("Invalid throughput value")



def latencyDistributionByThroughputGroup(df: pd.DataFrame):
  summary_table = df.groupby('load_type', observed=True).agg(
      mean=('value', 'mean'),
      std=('value', 'std'),
      p50=('value', lambda x: x.quantile(0.50)),
      p90=('value', lambda x: x.quantile(0.90)),
      p95=('value', lambda x: x.quantile(0.95)),
      p99=('value', lambda x: x.quantile(0.99)),
      error_rate=('is_error', lambda x: x[x == True].count() / x.count())
  )
  summary_table.loc["minimum", "min_load"] = 0
  summary_table.loc["minimum", "max_load"] = round(df.query("load_type == 'minimum'")["load_max"].mean())
  summary_table.loc["medium", "min_load"] = summary_table.loc["minimum", "max_load"]
  summary_table.loc["medium", "max_load"] = round(df.query("load_type == 'medium'")["load_max"].mean())
  summary_table.loc["large", "min_load"] = summary_table.loc["medium", "max_load"]
  summary_table.loc["large", "max_load"] = round(df.query("load_type == 'large'")["load_max"].mean())
  summary_table.loc["maximum", "min_load"] = summary_table.loc["large", "max_load"]
  summary_table.loc["maximum", "max_load"] = round(df.query("load_type == 'maximum'")["load_max"].mean())
  summary_table["load_range"] = summary_table.apply(lambda row: f"{row['min_load']}-{row['max_load']}", axis=1)
  ## remove underscore from index
  summary_table = summary_table.reset_index().rename(columns={
    "error_rate": "error rate",
    "load_type": "throughput",
    "load_range": "rps"
  })
  return summary_table