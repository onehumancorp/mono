#!/bin/bash
bazelisk run //srcs/cmd/ohc:ohc &
SERVER_PID=$!

cd srcs/frontend
npm install
npm run dev &
FRONTEND_PID=$!

sleep 10
