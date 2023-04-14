Modbus TCP Client Simulator
===========================
This is a simple command-line based Modbus TCP client simulator, useful for testing and simulating Modbus TCP devices. It allows you to perform various Modbus operations such as reading and writing coils and registers.

Features
--------
Perform Modbus TCP read and write operations
Support for signed and unsigned register values
Repeat operations at specified intervals
Easily configurable through command-line flags

Installation
------------
To build the program, make sure you have Go installed on your machine. Then, clone the repository and build the binary:

```bash
git clone https://github.com/torosalmonpink/modbus_client
cd modbus_client
go build -o modbus-client
```
This will create an executable called modbus-client.

Usage
-----
You can run the program using the generated binary. Use the following command to see available options:

```bash
./modbus-client --help
```

Example
-------
To perform a read operation for holding registers starting from address 0 with a count of 10:

```bash
./modbus-client -s 192.168.1.10 -p 502 -o read_holding_registers -u -start 0 -count 10
```

This will read holding registers from the Modbus server at IP address 192.168.1.10 on port 502. The -u flag indicates that the values should be treated as unsigned.

License
-------
This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.