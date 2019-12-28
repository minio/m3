#!/bin/bash
startVault="./vault server -config vault-config.json";
export VAULT_ADDR=http://127.0.0.1:8200
unsealVault="./vault operator init";

$startVault &
sleep 3
$unsealVault | grep -E -- "Unseal Key|Initial Root Token:" | while read -ra line
do
    string="${line[@]}";
    if [[ $string == *"Unseal Key"* ]]; then
      IFS=' ' read -ra unseal_token_array <<< "$string"
      echo "unsealing with token: ${unseal_token_array[3]}"
      ./vault operator unseal "${unseal_token_array[3]}"
    fi
    if [[ $string == *"Initial Root Token"* ]]; then
      IFS=' ' read -ra root_token_array <<< "$string"
      echo "root token: ${root_token_array[3]}"
      export VAULT_TOKEN="${root_token_array[3]}"
      ./vault auth enable approle
      ./vault secrets enable kv
    fi
done