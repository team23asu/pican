# Set up a prototype Pi

Use Raspberry Pi Imager application to put Raspberry Pi OS onto an SD card. 
Before writing the SD card you can click a gear icon and select the options that will enable SSH and auto-join Wifi so that you can connect remotely.

I chose a user named `team23` and gave it the password `we love gantt charts`. 
This should be changed if this thing spends time on the internet, but for development I will keep it this way out of convenience.

I booted the Pi, plugged into an HDMI monitor, keyboard and mouse.

# CAN HAT setup instructions

Follow all the setup instructions from Copperhill here to set up the CAN HAT:
https://www.waveshare.com/wiki/2-CH_CAN_HAT


# Listen to chair in normal operation

From here, you should be able to wire the Pi up to the Permobil chair communication bus in order to inspect CAN traffic.

With the chair plugged in in its normal state, run jumper wires from CAN_H, CAN_L and GND to the R-Net connector. You can use either CAN0 or CAN1.

```
candump can0
# or
candump can1

# Ctrl-C to break
```

Log the messages to a file while watching them by:

```
candump -L can0 | tee "candump-$(date -Iseconds).log"
```

## Goals:

* make a note of the Joystick ID. Look for the `0x02000_00#XxYy` messages where _ is a hexadecimal digit corresponding to the Joystick Module's ID and `Xx` and `Yy` are hex numbers between -100 and 100 decimal.
* log interesting messages to a file, from chair power-on to joystick operation frames


# (optional) Run Anthony's code to filter/rewrite joystick messages

Additionally, you can install Visual Studio Code, Git and Golang in order to run the code in this repository:

```
sudo apt update
sudo apt install can-utils code git golang

mkdir workspace && cd workspace
git clone https://github.com/team23asu/pican
code pican/
```

Also I made a note of my processor architecture, `armv7` in case for some reason it's different between our machines. `cat /proc/cpuinfo` shows 4 ARMv7 cores.


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
