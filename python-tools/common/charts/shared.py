import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import matplotlib.dates as mdates
import scipy.stats as stats


def normalizeTime(df: pd.DataFrame):
  start = pd.Timestamp(year=2000, month=1, day=1, tz="UTC")
  difference = df["time"].min() - start
  df["time"] = df["time"] - difference
  return df

def wassersteinDistance(x):
  real = x.query("source == 'real'")
  mock = x.query("source == 'mock'")
  meanReal = real["value"].mean()
  if len(real) == 0 or len(mock) == 0:
    return None
  return stats.wasserstein_distance(real["value"], mock["value"]) / meanReal

