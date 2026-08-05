#!/bin/bash

available_ram_mb=$(awk '/MemAvailable/ {print int($2 / 1024)}' /proc/meminfo)

# Ubuntu
# cpu_info=$(top -bn1 | grep "Cpu")
# available_cpu_percentage=$(echo "$cpu_info" | awk '{for (i=1; i<=NF; i++) if ($i == "id,") print $(i-1)}')

# Alpine
cpu_info=$(top -bn1 | grep "CPU:")
available_cpu_percentage=$(echo "$cpu_info" | awk '{for (i=1; i<=NF; i++) if ($i == "idle") print $(i-1)}')
available_cpu_percentage=$(echo "$available_cpu_percentage" | sed 's/%//')


echo "${available_ram_mb}"
echo "${available_cpu_percentage}"

onegate vm update --data "FREE_MEM=\"${available_ram_mb}\""
onegate vm update --data "FREE_CPU=\"${available_cpu_percentage}\""