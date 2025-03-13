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

