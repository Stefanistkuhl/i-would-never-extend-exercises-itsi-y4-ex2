#!/usr/bin/env fish

# PCAP capture script with auto-naming
# Creates 10 example captures in correct format

set -l CAPTURE_DIR ./captures/incoming
set -l DURATION 10
set -l INTERFACE wlan0

mkdir -p $CAPTURE_DIR

set -l HOSTNAMES SRV1 SRV2
set -l SCENARIOS baseline http https ssh dns attack normal test forensics

for i in (seq 1 10)
    # Use different variable names (not hostname)
    set -l host $HOSTNAMES[(random 1 (count $HOSTNAMES))]
    set -l scenario $SCENARIOS[(random 1 (count $SCENARIOS))]
    set -l now (date +%Y%m%d_%H%M%S)
    set -l filename "$CAPTURE_DIR/{$host}_{$scenario}_{$now}.pcap"

    echo "Capturing #$i: $filename"
    sudo tcpdump -i $INTERFACE -w $filename -G $DURATION 'tcp or udp' 2>/dev/null
    sleep 2
    echo "Saved: $filename"
    echo ""
end

echo "Done! 10 captures created"
ls -lh $CAPTURE_DIR
