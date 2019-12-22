
./m3 cluster sc add --name my-dc-rack-1

./m3 cluster nodes add --name node-1 --k8s_label m3cluster-worker --volumes /mnt/disk{1...4}
./m3 cluster nodes add --name node-2 --k8s_label m3cluster-worker2 --volumes /mnt/disk{1...4}
./m3 cluster nodes add --name node-3 --k8s_label m3cluster-worker3 --volumes /mnt/disk{1...4}
./m3 cluster nodes add --name node-4 --k8s_label m3cluster-worker4 --volumes /mnt/disk{1...4}

./m3 cluster nodes assign --storage_cluster my-dc-rack-1 --node node-1
./m3 cluster nodes assign --storage_cluster my-dc-rack-1 --node node-2
./m3 cluster nodes assign --storage_cluster my-dc-rack-1 --node node-3
./m3 cluster nodes assign --storage_cluster my-dc-rack-1 --node node-4

./m3 cluster sc sg add --storage_cluster my-dc-rack-1 --name group-1

#./m3 tenant add "Acme Inc." --short_name acme --admin_name="Your Name" --admin_email="your@mail.com"
