#!/bin/bash

go build
echo "Run:"
echo "./app -a $MONGODB_HOST -p 27017 -m mountpoint/"
echo ""
bash
