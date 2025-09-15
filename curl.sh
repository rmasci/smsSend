#!/bin/bash

curl -X POST \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=${CLIENT_ID}" \
  -d "client_secret=${CLIENT_SECRET}" \
  -d "grant_type=client_credentials" \
  -d "scope=https://communication.azure.com/.default" \
  "https://login.microsoftonline.com/${TENANT_ID}/oauth2/v2.0/token"
