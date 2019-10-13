create schema provisioning;

create table provisioning.tenants
(
    id serial not null
        constraint tenants_pk
            primary key,
    name varchar(256) not null,
    short_name varchar(256) not null
);

alter table provisioning.tenants owner to postgres;

create table provisioning.storage_clusters
(
    id serial not null
        constraint storage_clusters_pk
            primary key,
    name varchar(256)
);

alter table provisioning.storage_clusters owner to postgres;

create table provisioning.tenants_storage_clusters
(
    tenant_id integer not null
        constraint tenants_storage_clusters_tenants_id_fk
            references provisioning.tenants,
    storage_cluster_id integer not null
        constraint tenants_storage_clusters_storage_clusters_id_fk
            references provisioning.storage_clusters,
    port integer not null,
    service_name varchar(64) not null
);

alter table provisioning.tenants_storage_clusters owner to postgres;

create table provisioning.nodes
(
    id serial not null
        constraint nodes_pk
            primary key,
    name varchar(256),
    k8s_label varchar(256)
);

alter table provisioning.nodes owner to postgres;

create table provisioning.storage_clusters_nodes
(
    storage_cluster_id integer not null
        constraint storage_clusters_nodes_storage_clusters_id_fk
            references provisioning.storage_clusters,
    node_id integer not null
        constraint storage_clusters_nodes_nodes_id_fk
            references provisioning.nodes
);

alter table provisioning.storage_clusters_nodes owner to postgres;

create table provisioning.node_volumes
(
    id serial not null
        constraint node_volumes_pk
            primary key,
    node_id integer not null,
    mount_path varchar(256)
);

alter table provisioning.node_volumes owner to postgres;

