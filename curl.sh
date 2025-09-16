#!/bin/bash

token=$(curl -X POST \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=${CLIENT_ID}" \
  -d "client_secret=${CLIENT_SECRET}" \
  -d "grant_type=client_credentials" \
  -d "scope=https://communication.azure.com/.default" \
  "https://login.microsoftonline.com/${TENANT_ID}/oauth2/v2.0/token")


curl -X POST \
  "https://${RESOURCE_NAME}.communication.azure.com/sms?api-version=2021-03-07" \
  -H "Authorization: Bearer ${token}" \
  -H "Content-Type: application/json" \
  -d '{
        "from": "${FROM}",
        "smsRecipients": [
          {
            "to": "${TO}"
          }
        ],
        "message": "Hello from Azure Communication Services via curl!",
        "smsSendOptions": {
          "enableDeliveryReport": true,
          "tag": "test-message"
        }
      }'
