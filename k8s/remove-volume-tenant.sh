#!/bin/bash

TENANT_NAME="${1:-tenant-1}"

docker exec -it m3cluster-worker rm -Rf /mnt/disk1/$TENANT_NAME
docker exec -it m3cluster-worker rm -Rf /mnt/disk2/$TENANT_NAME
docker exec -it m3cluster-worker rm -Rf /mnt/disk3/$TENANT_NAME
docker exec -it m3cluster-worker rm -Rf /mnt/disk4/$TENANT_NAME

docker exec -it m3cluster-worker2 rm -Rf /mnt/disk1/$TENANT_NAME
docker exec -it m3cluster-worker2 rm -Rf /mnt/disk2/$TENANT_NAME
docker exec -it m3cluster-worker2 rm -Rf /mnt/disk3/$TENANT_NAME
docker exec -it m3cluster-worker2 rm -Rf /mnt/disk4/$TENANT_NAME

docker exec -it m3cluster-worker3 rm -Rf /mnt/disk1/$TENANT_NAME
docker exec -it m3cluster-worker3 rm -Rf /mnt/disk2/$TENANT_NAME
docker exec -it m3cluster-worker3 rm -Rf /mnt/disk3/$TENANT_NAME
docker exec -it m3cluster-worker3 rm -Rf /mnt/disk4/$TENANT_NAME

docker exec -it m3cluster-worker4 rm -Rf /mnt/disk1/$TENANT_NAME
docker exec -it m3cluster-worker4 rm -Rf /mnt/disk2/$TENANT_NAME
docker exec -it m3cluster-worker4 rm -Rf /mnt/disk3/$TENANT_NAME
docker exec -it m3cluster-worker4 rm -Rf /mnt/disk4/$TENANT_NAME