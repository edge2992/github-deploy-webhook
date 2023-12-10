#!/bin/bash

set -e

cd /home/edge2992/ghq/github.com/edge2992/diary

git pull origin main

docker compose up --build -d

echo "Deployment completed successfully."
