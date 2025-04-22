
# Simulator linked-list
This is the simulator used in this thesis. It is implemented using a simple linked list. It can be started, given a function definition using
```
go run . <json_definition_name> (e.g. go-factorial)
```

See `go-factorial.json` for an example.

Each function instance for which a function profile is provided can then be reached at `http://localhost:8089?name=<function_name>-<replica-number>`