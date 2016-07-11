#!/usr/bin/env bash
sudo apt-get update
sudo apt-get upgrade

# install openjdk-7 
#sudo apt-get purge openjdk*
#sudo apt-get -y install openjdk-7-jdk
 
# install golang
cd /usr/local
sudo wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.6.2.linux-amd64.tar.gz

# add golang to path
echo 'export PATH=$PATH:/usr/local/go/bin' >>/home/vagrant/.bash_profile
export PATH=$PATH:/usr/local/go/bin

# go is working
echo go help...
go help

# install git
apt-get -y install git

# creating new workspace directory at home directory
mkdir /home/vagrant/workspace
cd /home/vagrant/workspace

# setting GOPATH...
export GOPATH=/home/vagrant/workspace/gostuff

# getting pz-uuidgen and trying to build it
go get github.com/venicegeo/pz-uuidgen
go install github.com/venicegeo/pz-uuidgen

# start the app
cd /home/vagrant/workspace/gostuff/bin
echo List of built executables:
ls -la

echo starting pz-uuidgen...
./pz-uuidgen
