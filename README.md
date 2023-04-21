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

## License

This project is released under the MIT License. See the LICENSE file for details.