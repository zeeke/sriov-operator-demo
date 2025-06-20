# SR-IOV Network Operator Demo

Generate SR-IOV Network Operator resources for demostration purpose.

## Installation

The tool can be installed via
```
go install github.com/zeeke/sriov-operator-demo@latest
```

## Usage

List all the available scenarios
```
sriov-operator-demo list
```

Generate yaml resources for a scenario
```
sriov-operator-demo --scenario intel-demo
```


Apply generated resources
```
sriov-operator-demo --scenario intel-demo | oc apply -f -
```

## Available Scenarios

The following scenarios are currently supported:

<!-- AUTO-GENERATED-SCENARIOS-START -->
- `intel-nics`
- `mellanox-nics`
- `switchdev`
<!-- AUTO-GENERATED-SCENARIOS-END -->

## Updating Documentation

To update the README.md with the current list of available scenarios, run:

```bash
make update-doc
```

This will automatically regenerate the "Available Scenarios" section above.

