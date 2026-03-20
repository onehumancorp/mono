#!/bin/bash
cd srcs/frontend
npm run dev -- --host 127.0.0.1 --port 8081 > frontend_dev.log 2>&1 &
echo "Started frontend."
