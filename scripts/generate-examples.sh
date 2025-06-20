#!/bin/bash -x

exec='go run main.go'

declare -A cmds
#cmds['intel-nics.yaml']="$exec -s intel-nics"
#cmds['mellanox-nics.yaml']="$exec -s mellanox-nics"
cmds['dpdk-tap.yaml']="$exec -s dpdk-tap"

# cmds['switchdev.yaml']="$exec -s switchdev"

for key in ${!cmds[@]}; do
    echo $key
    cmd=${cmds[$key]}
    echo "# Generated with"
    echo "# $cmd" > ./examples/${key}
    echo >> ./examples/$key
    $cmd >> ./examples/$key
done
