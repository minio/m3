TENANT_NAME="${1:-tenant-1}"

docker exec -it m3cluster-worker mkdir -p /mnt/disk1/$TENANT_NAME
docker exec -it m3cluster-worker mkdir -p /mnt/disk2/$TENANT_NAME
docker exec -it m3cluster-worker mkdir -p /mnt/disk3/$TENANT_NAME
docker exec -it m3cluster-worker mkdir -p /mnt/disk4/$TENANT_NAME

docker exec -it m3cluster-worker2 mkdir -p /mnt/disk1/$TENANT_NAME
docker exec -it m3cluster-worker2 mkdir -p /mnt/disk2/$TENANT_NAME
docker exec -it m3cluster-worker2 mkdir -p /mnt/disk3/$TENANT_NAME
docker exec -it m3cluster-worker2 mkdir -p /mnt/disk4/$TENANT_NAME

docker exec -it m3cluster-worker3 mkdir -p /mnt/disk1/$TENANT_NAME
docker exec -it m3cluster-worker3 mkdir -p /mnt/disk2/$TENANT_NAME
docker exec -it m3cluster-worker3 mkdir -p /mnt/disk3/$TENANT_NAME
docker exec -it m3cluster-worker3 mkdir -p /mnt/disk4/$TENANT_NAME

docker exec -it m3cluster-worker4 mkdir -p /mnt/disk1/$TENANT_NAME
docker exec -it m3cluster-worker4 mkdir -p /mnt/disk2/$TENANT_NAME
docker exec -it m3cluster-worker4 mkdir -p /mnt/disk3/$TENANT_NAME
docker exec -it m3cluster-worker4 mkdir -p /mnt/disk4/$TENANT_NAME