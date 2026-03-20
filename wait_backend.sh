#!/bin/bash
echo "Waiting for backend to start..."
while ! curl -s http://127.0.0.1:8080 > /dev/null; do
  sleep 1
done
echo "Backend started!"
