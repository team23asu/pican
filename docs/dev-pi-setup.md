# Notes on setting up the Raspberry Pi

I used Raspberry Pi Imager application to put Raspberry Pi OS onto an SD card. 
Before writing the SD card you can click a gear icon and select the options that will enable SSH and auto-join your Wifi.

I chose a user named `team23` and gave it the password `we love gantt charts`. 
This should be changed if this thing spends time on the internet, but for development I will keep it this way out of convenience.

I booted the Pi, plugged into an HDMI monitor, keyboard and mouse.

Once the graphical desktop loads, I open the terminal and install things:

```
sudo apt update
sudo apt install can-utils code git golang
# optional, for developing the gui using ebiten
sudo apt install libc6-dev libglu1-mesa-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config

mkdir workspace && cd workspace
git clone https://github.com/team23asu/pican
code pican/
```

Also I made a note of my processor architecture, `armv7` in case for some reason it's different between our machines. `cat /proc/cpuinfo` shows 4 ARMv7 cores.

Now I need virtual CAN interfaces so I can develop against them.

```
# enable virtual CAN "vcan" kernel module
sudo modprobe vcan
# add "vcan0" and "vcan1"
sudo ip link add dev vcan0 type vcan
sudo ip link add dev vcan1 type vcan
# enable "vcan0" and "vcan1"
sudo ip link set up vcan0
sudo ip link set up vcan1
```

Verify your interfaces are usable:
In one terminal:
```
# output all messages on "vcan0"
candump vcan0
```

In a second terminal:
```
# send a test message
cansend vcan0 123#0000000000000000
```

You should see an output in the first terminal window.

Other useful commands

```
# store incoming messages on vcan0 in logfile format
candump -L "vcan0" > vcan0messages.log

# then, later, replay the same messages for testing
canplayer -I vcan0messages.log
```
