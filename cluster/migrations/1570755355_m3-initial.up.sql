create schema provisioning;

create table provisioning.tenants
(
    id uuid not null
        constraint tenants_pk
            primary key,
    name varchar(256) not null,
    short_name varchar(256) not null
);

alter table provisioning.tenants owner to postgres;

create table provisioning.storage_groups
(
    id uuid not null
        constraint storage_groups_pk
            primary key,
    name varchar(256),
    num serial not null
);

alter table provisioning.storage_groups owner to postgres;

create table provisioning.tenants_storage_groups
(
    tenant_id uuid not null
        constraint tenants_storage_groups_tenants_id_fk
            references provisioning.tenants,
    storage_group_id uuid not null
        constraint tenants_storage_groups_storage_groups_id_fk
            references provisioning.storage_groups,
    port integer not null,
    service_name varchar(64) not null
);

alter table provisioning.tenants_storage_groups owner to postgres;

create table provisioning.nodes
(
    id uuid not null
        constraint nodes_pk
            primary key,
    name varchar(256),
    k8s_label varchar(256)
);

alter table provisioning.nodes owner to postgres;

create table provisioning.storage_clusters
(
    id uuid not null
        constraint storage_clusters_pk
            primary key,
    name varchar(256)
);

alter table provisioning.storage_clusters owner to postgres;

create table provisioning.storage_clusters_groups
(
    storage_cluster_id uuid not null
        constraint storage_clusters_groups_storage_clusters_id_fk
            references provisioning.storage_clusters,
    storage_group_id uuid not null
        constraint storage_clusters_groups_storage_groups_id_fk
            references provisioning.storage_groups
);

alter table provisioning.storage_clusters_groups owner to postgres;

create table provisioning.storage_cluster_nodes
(
    storage_cluster_id uuid not null
        constraint storage_cluster_nodes_storage_clusters_id_fk
            references provisioning.storage_clusters,
    node_id uuid not null
        constraint storage_cluster_nodes_nodes_id_fk
            references provisioning.nodes,
    k8s_label varchar(256)
);

alter table provisioning.storage_cluster_nodes owner to postgres;

create table provisioning.node_volumes
(
    id uuid not null
        constraint node_volumes_pk
            primary key,
    node_id uuid not null
        constraint node_volumes_nodes_id_fk
            references provisioning.nodes,
    mount_path varchar(256)
);

alter table provisioning.node_volumes owner to postgres;


