
import seaborn as sns
import os
import common.lazy as lazy
import common.charts.shared as shared
import common.charts.validation as chars
import matplotlib.pyplot as plt
sns.set_theme()

for func in ['go-factorial', 'go-io', 'go-string-reversal']:
  functionName = f"{func}-validation"
  benchmarkName = f"{func}-validation-1"
  mockBenchmarkName = f"{func}-mock-validation-1"
  outputFolder = f"output/{func}"
  if not os.path.exists(outputFolder):
    os.makedirs(outputFolder)
  httpRequestDuration = shared.normalizeTime(lazy.getHttRequestStats(benchmarkName, "32"))
  httpRequestDuration["source"] = "real"
  mockHttpRequestDuration = shared.normalizeTime(lazy.getHttRequestStats(mockBenchmarkName, "32"))
  mockHttpRequestDuration["source"] = "mock"
  httpReqSending = shared.normalizeTime(lazy.getHttpReqSending(mockBenchmarkName, "32"))
  fig, axes = chars.validationComparisonPlot(httpRequestDuration, mockHttpRequestDuration, httpReqSending)
  plt.savefig(f"{outputFolder}/{func}-validation-comparison.png")
  plt.savefig(f"{outputFolder}/{func}-validation-comparison.pdf")
  plt.close()
  print(f"Saved {outputFolder}/{func}-validation-comparsion.pdf")
  print(f"Done for {func}")
