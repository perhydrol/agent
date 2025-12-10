#!/bin/bash

echo "Stopping Development Environment..."
podman rm -f insurai-mysql insurai-redis
echo "Done."
