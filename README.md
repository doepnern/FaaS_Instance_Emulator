# Bachelorarbeit_3s_FaaS_Code

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

