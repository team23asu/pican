# Team 23 Pi + CAN test harness

This repository is developed for the ASU project team Team 23, working on a Collision Preventive Wheelchair.

To run the graphical demo, you must have a few things installed.

## Install dependencies (Raspbian/Debian/Ubuntu Linux)

```
sudo apt update
sudo apt install can-utils golang
sudo apt install libc6-dev libglu1-mesa-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config

```

## Install dependencies (Mac)

```
brew install golang
```

## Run the demo from the repository directory

```
go run cmd/demo1/main.go
```
