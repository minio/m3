#!/bin/bash

#!/bin/bash

TENANT_NAME="${1:-tenant-1}"
WORKER=""
for i in {1..4}; do
  docker exec -it m3cluster-worker$WORKER rm -Rf /mnt/disk{1..8}/"$TENANT_NAME"
  Y=1
  WORKER="$(($i + $Y))"
done
