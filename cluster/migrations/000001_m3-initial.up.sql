create schema provisioning;

create table provisioning.tenants
(
    id         uuid         not null
        constraint tenants_pk
            primary key,
    name       varchar(256) not null,
    short_name varchar(256) not null
);


create table provisioning.storage_groups
(
    id   uuid   not null
        constraint storage_groups_pk
            primary key,
    name varchar(256),
    num  serial not null
);


create table provisioning.tenants_storage_groups
(
    tenant_id        uuid        not null
        constraint tenants_storage_groups_tenants_id_fk
            references provisioning.tenants,
    storage_group_id uuid        not null
        constraint tenants_storage_groups_storage_groups_id_fk
            references provisioning.storage_groups,
    port             integer     not null,
    service_name     varchar(64) not null
);


create table provisioning.nodes
(
    id        uuid not null
        constraint nodes_pk
            primary key,
    name      varchar(256),
    k8s_label varchar(256)
);


create table provisioning.storage_clusters
(
    id   uuid not null
        constraint storage_clusters_pk
            primary key,
    name varchar(256)
);


create table provisioning.storage_clusters_groups
(
    storage_cluster_id uuid not null
        constraint storage_clusters_groups_storage_clusters_id_fk
            references provisioning.storage_clusters,
    storage_group_id   uuid not null
        constraint storage_clusters_groups_storage_groups_id_fk
            references provisioning.storage_groups
);


create table provisioning.storage_cluster_nodes
(
    storage_cluster_id uuid not null
        constraint storage_cluster_nodes_storage_clusters_id_fk
            references provisioning.storage_clusters,
    node_id            uuid not null
        constraint storage_cluster_nodes_nodes_id_fk
            references provisioning.nodes,
    k8s_label          varchar(256)
);


create table provisioning.node_volumes
(
    id         uuid not null
        constraint node_volumes_pk
            primary key,
    node_id    uuid not null
        constraint node_volumes_nodes_id_fk
            references provisioning.nodes,
    mount_path varchar(256)
);

--  Table to store Disks attached to a node and their mount points

create table provisioning.disks
(
    id          uuid not null
        constraint disks_pk
            primary key,
    node_id     uuid
        constraint disks_nodes_id_fk
            references provisioning.nodes,
    mount_point varchar(512),
    capacity    bigint
);

comment on column provisioning.disks.capacity is 'Capacity in bytes';

--  Table to store sessions of a <tenant>.user

create table provisioning.sessions
(
    id          varchar(256) not null
        constraint sessions_pk
            primary key,       -- session id as rand string 
    tenant_id   uuid not null, -- user's tenant's id
    user_id     uuid not null, -- user id of the user who initiated the session
    occurred_at timestamp with time zone default now() not null, -- first timestamp of the session
    last_event  timestamp with time zone default now()           -- stores last event's timestamp within this session
);


