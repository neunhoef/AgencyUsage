#!/bin/bash
cd AgencyCommTest
go build AgencyCommTest.go
cd ..

cd discoveryStandAlone
go build discoveryStandAlone.go
cd ..

cd discovery
go build discovery.go
cd ..

cd discovery.docker
cp ../discovery/discovery .
strip ./discovery
docker build -t neunhoef/discovery .
docker push neunhoef/discovery
cd ..

