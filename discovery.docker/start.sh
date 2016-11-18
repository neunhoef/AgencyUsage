#!/bin/bash
env
exec /discovery $HOST:$PORT0 agency.marathon.l4lb.thisdcos.directory:8529
