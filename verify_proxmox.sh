#!/bin/bash

# Check if required arguments are provided
if [ $# -ne 2 ]; then
  echo "Usage: $0 <host> <path-to-private-key>"
  echo "Example: $0 192.168.1.100 ~/.ssh/id_ed25519"
  exit 1
fi

HOST=$1
PRIVATE_KEY_PATH=$2

echo "Verifying SSH access for dannyvelasquez user..."
if ssh -o BatchMode=yes -o StrictHostKeyChecking=no -i "$PRIVATE_KEY_PATH" "dannyvelasquez@$HOST" "echo 'SSH access successful for dannyvelasquez'"; then
  echo "dannyvelasquez SSH access verified"
else
  echo "Error: Unable to verify SSH access for dannyvelasquez"
fi

echo "Verifying SSH access and privileges for terraform user..."
if ssh -o BatchMode=yes -o StrictHostKeyChecking=no -i "$PRIVATE_KEY_PATH" "terraform@$HOST" "sudo pvesm apiinfo"; then
  echo "terraform SSH access and privileges verified"
else
  echo "Error: Unable to verify SSH access or privileges for terraform user"
fi
