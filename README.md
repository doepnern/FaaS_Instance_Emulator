# Bachelors thesis 3S FaaS Emulator Code
This repository contains the code for my bachelors thesis at the Scalable Software Systems Reseach Group at TU Berlin.

### Contents
The [benchmarking client](./benchmarks) is used to orchestrate the benchmarks in this thesis. The implemented FaasEmulator can be found in [simulator-linked-list](./simulator-linked-list). An alternative emulator with a more flexible api is implemented in [simulator-sqlite](./simulator-sqlite), it was not evaluated as part of the thesis though, as the added complexity is not needed for the proposed architecture. The benchmarked functions are implemented in [function](./functions).
## Getting started

- install node and run `npm i`
- install bun
- set up docker to run in isolated resource slice
- Install prometheus and cadvisor on server
- make sure prometheus is running and collecting cadvisor metrics
- set correct prometheus url in `benchmarks/.env`

### Limiting docker resources on Ubuntu

[Guide](https://gist.github.com/ryderstorm/61943436751cb2d848202cda0ad26dd2)

##@ Note

Memory usage is caluculcated according to [Docker docs](https://docs.docker.com/reference/cli/docker/container/stats/) assuming the benchmark is run on docker version > 19.03 and a host using cgroup v2.
