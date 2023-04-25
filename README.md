# sml2mqtt

`sml2mqtt` is a Go-based command line tool that reads SML (Smart Message Language) data from a file or device, extracts specific sensor values, and publishes them to an MQTT broker. This tool can be useful for monitoring energy meters or other devices that use the SML format.

## Features

* Read SML data from a file or device
* Supports multiple sensor values and extracts them based on user-defined criteria
* Publish extracted sensor values to an MQTT broker and allows to auto discover them by home assistant

## Prerequisites

* Go (version 1.16 or higher)
* MQTT broker (e.g., Mosquitto)

## Installation

1. Clone this repository:
  ```bash
  git clone https://github.com/mfmayer/sml2mqtt.git
  ```

2. Change to the project directory:
  ```bash
  cd sml2mqtt
  ```

3. Build the binary:
  ```bash
  go build -o sml2mqtt
  ```

4. Move the binary to your desired location (optional):
  ```bash
  sudo mv sml2mqtt /usr/local/bin/
  ```

## Usage

```bash
sml2mqtt [options]
```

Reads file and published specified values to given broker.

### Options:

* **`-file <path>`**: Path to the file with sensor values.
* **`-broker <url>`**: URL to the MQTT broker.
* **`-value <sensorValue>`**: Value's OBIS code, ValueName, DeviceClass, UnitOfMeasure and optional CorrectionFactor. Format: ObisCode,ValueName,DeviceClass,UnitOfMeasure[,CorrectionFactor] (e.g., "1.0.1.8.0,Energy,energy,kWh[,0.001]"). Multiple sensor values can be specified.

## Example

```bash
sml2mqtt -file /path/to/sml/file -broker mqtt://localhost:1883 -value "1.0.1.8.0,Energy,energy,kWh,0.001"
```

## Serial Device Files

When you want to read the SML from a serial device file, ensure to appropriately configure the serial device beforehand. For example the following command configures the serial interface `/dev/ttyUSB0` with the following settings: 9600 baud rate, 8 data bits, 1 stop bit, no parity checking, and in raw mode.: 

```bash
sudo stty -F /dev/ttyUSB0 9600 cs8 -cstopb -parenb raw
```

## UDEV Rules

When you want to read the SML from multiple serial device files (e.g. multiple UART USB devices), it might be helpful to name dem specifally based on the USB port they're plugged in. 

To find specific information that you want to use in UDEV configurations you can use the command `udevadm info -a -n /dev/ttyUSB0`

To e.g. find the USB port the device is plugged in execute: `udevadm info -a -n /dev/ttyUSB0 | grep devpath` (or `idProduct` for the product ID)

So exemplary UDEV rules for two serial devices to read out electricity meters could be (`/etc/udev/rules.d/99-usb-serial.rules`):

```
SUBSYSTEM=="tty", ATTRS{idVendor}=="1a86", ATTRS{idProduct}=="7523", ATTRS{devpath}=="1.2", SYMLINK+="emeter0"

SUBSYSTEM=="tty", ATTRS{idVendor}=="1a86", ATTRS{idProduct}=="7523", ATTRS{devpath}=="1.3", SYMLINK+="emeter1"
```

## SystemD Service

In order to create a SystemD service you freate a SystemD service file

```bash
sudo nano /etc/systemd/system/sml2mqtt_emeter0.service
```

with the following content:

```
[Unit]
Description=SML to MQTT Service
After=network.target

[Service]
Type=simple
User=sml2mqtt
WorkingDirectory=/var/lib/sml2mqtt
ExecStartPre=/usr/bin/stty -F /dev/emeter0 9600 cs8 -cstopb -parenb raw
ExecStart=/usr/local/bin/sml2mqtt --file /dev/emeter0 --broker tcp://horst:1883 --value 1.0.1.8.0,Energy,energy,kWh,0.001 --value 1.0.16.7.0,Power,power,W
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

This will read out enery (OBIS `1-1:1.8.0`) and power (OBIS `1-0:16.7.0`) from `/dev/emeter0` and provide it to give MQTT broker with the topics:

* `homeassistant/sensor/emeter0_Energy/config` - to register the entity with home assistant
* `homeassistant/sensor/emeter0_Energy/state` - to publish/update the value

The config topic is only published at startup and then from time to time every few minutes.

After creating the service file, don't forget to update the systemd configuration by reloading its deamon:

```bash
sudo systemctl daemon-reload
```

And finally start and enable the service (if it shall be restarted after reboot)

```bash
sudo systemctl start sml2mqtt_emeter0
sudo systemctl enable sml2mqtt_emeter0
```

## License

This project is released under the MIT License. See the LICENSE file for details.