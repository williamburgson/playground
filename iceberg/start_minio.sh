#!/bin/bash 
set -eu
source .env
minio server .
