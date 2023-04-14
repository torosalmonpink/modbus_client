package main

import (
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/goburrow/modbus"
	"github.com/spf13/pflag"
)

// ModbusArgs holds the parsed command-line arguments
type ModbusArgs struct {
	Server    string
	Port      uint
	UnitID    byte
	Operation string
	Start     uint16
	Count     uint16
	Value     uint16
	Values    []uint16
	Repeat    int
	Interval  int
	Unsigned  bool
}

// parseFlags parses the command-line arguments and returns a ModbusArgs struct
func parseFlags() *ModbusArgs {
	args := &ModbusArgs{}

	pflag.StringVarP(&args.Server, "server", "s", "", "The IP address or hostname of the Modbus TCP server.")
	pflag.UintVarP(&args.Port, "port", "p", 502, "The port number of the Modbus TCP server.")
	pflag.Uint8VarP(&args.UnitID, "unitid", "d", 1, "The unit id of the Modbus TCP server.")
	pflag.StringVarP(&args.Operation, "operation", "o", "", "The operation to perform. \nread_coils/read_discrete_inputs/read_holding_registers/read_input_registers\nwrite_single_coil/write_single_register/write_multiple_coils/write_multiple_registers")
	pflag.IntVarP(&args.Repeat, "repeat", "r", 1, "The number of times the operation should be repeated. If set to 0, repeat until interrupted.")
	pflag.IntVarP(&args.Interval, "interval", "i", 1000, "The interval (in milliseconds) between operation repeats.")
	pflag.BoolVarP(&args.Unsigned, "unsigned", "u", false, "Interpret read/write values as unsigned integers.")
	pflag.Uint16VarP(&args.Start, "start", "", 0, "The starting address for read or write operations.")
	pflag.Uint16VarP(&args.Count, "count", "", 1, "The number of registers to read.")
	var valueStr string
	pflag.StringVarP(&valueStr, "value", "", "0", "The value for single write operations.")
	var values []string
	pflag.StringSliceVarP(&values, "values", "", nil, "The comma-separated values for multiple write operations. Example: 1,2,3")

	pflag.Parse()

	// Validate server address
	if args.Server == "" {
		log.Fatal("Server address is required")
	}

	// Conditionally parse the value based on the --unsigned flag
	if args.Unsigned {
		value, err := strconv.ParseUint(valueStr, 10, 16)
		if err != nil {
			log.Fatalf("Invalid value: %s", valueStr)
		}
		args.Value = uint16(value)
	} else {
		value, err := strconv.ParseInt(valueStr, 10, 16)
		if err != nil {
			log.Fatalf("Invalid value: %s", valueStr)
		}
		args.Value = uint16(value & 0xFFFF)
	}

	// Convert the values from []string to []uint16
	args.Values = make([]uint16, len(values))
	for i, valueStr := range values {
		if args.Unsigned {
			value, err := strconv.ParseUint(valueStr, 10, 16)
			if err != nil {
				log.Fatalf("Invalid value in 'values': %s", valueStr)
			}
			args.Values[i] = uint16(value)
		} else {
			value, err := strconv.ParseInt(valueStr, 10, 16)
			if err != nil {
				log.Fatalf("Invalid value in 'values': %s", valueStr)
			}
			args.Values[i] = uint16(value & 0xFFFF)
		}
	}

	return args
}

// main is the entry point for the Modbus TCP client simulator
func main() {
	args := parseFlags()

	// Connect to the Modbus server
	handler, client := createModbusClient(args.Server, args.Port, args.UnitID)
	defer handler.Close()

	// Execute the requested operation
	switch args.Operation {
	case "read_coils":
		performReadOperation(client, modbus.FuncCodeReadCoils, args.Start, args.Count, args.Repeat, args.Interval, args.Unsigned)
	case "read_discrete_inputs":
		performReadOperation(client, modbus.FuncCodeReadDiscreteInputs, args.Start, args.Count, args.Repeat, args.Interval, args.Unsigned)
	case "read_holding_registers":
		performReadOperation(client, modbus.FuncCodeReadHoldingRegisters, args.Start, args.Count, args.Repeat, args.Interval, args.Unsigned)
	case "read_input_registers":
		performReadOperation(client, modbus.FuncCodeReadInputRegisters, args.Start, args.Count, args.Repeat, args.Interval, args.Unsigned)
	case "write_single_coil":
		writeSingleCoil(client, args.Start, args.Value, args.Repeat, args.Interval)
	case "write_single_register":
		writeSingleRegister(client, args.Start, args.Value, args.Repeat, args.Interval)
	case "write_multiple_coils":
		writeMultipleCoils(client, args.Start, args.Values, args.Repeat, args.Interval)
	case "write_multiple_registers":
		writeMultipleRegisters(client, args.Start, args.Values, args.Repeat, args.Interval)
	default:
		log.Fatalf("Invalid operation: %s", args.Operation)
	}
}

// createModbusClient creates a Modbus TCP client and connects to the server
func createModbusClient(server string, port uint, unitid uint8) (*modbus.TCPClientHandler, modbus.Client) {
	// Validate the server address
	addr := net.JoinHostPort(server, strconv.FormatUint(uint64(port), 10))
	handler := modbus.NewTCPClientHandler(addr)
	handler.SlaveId = byte(unitid)
	client := modbus.NewClient(handler)
	return handler, client
}

// performReadOperation is a helper function for read operations
func performReadOperation(client modbus.Client, functionCode byte, start uint16, count uint16, repeat int, interval int, unsigned bool) {
	for i := 0; repeat <= 0 || i < repeat; i++ {
		var response []byte
		var err error

		switch functionCode {
		case modbus.FuncCodeReadCoils:
			response, err = client.ReadCoils(start, count)
		case modbus.FuncCodeReadDiscreteInputs:
			response, err = client.ReadDiscreteInputs(start, count)
		case modbus.FuncCodeReadHoldingRegisters:
			response, err = client.ReadHoldingRegisters(start, count)
		case modbus.FuncCodeReadInputRegisters:
			response, err = client.ReadInputRegisters(start, count)
		}

		if err != nil {
			log.Printf("Error during read operation: %v", err)
		} else {
			if unsigned {
				values := make([]uint16, count)
				for i := 0; i < len(response); i += 2 {
					values[i/2] = binary.BigEndian.Uint16(response[i : i+2])
				}
				log.Printf("Read response (unsigned): %v", values)
			} else {
				values := make([]int16, count)
				for i := 0; i < len(response); i += 2 {
					values[i/2] = int16(binary.BigEndian.Uint16(response[i : i+2]))
				}
				log.Printf("Read response (signed): %v", values)
			}
		}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

// writeSingleCoil writes a single coil to the Modbus server
func writeSingleCoil(client modbus.Client, address uint16, value uint16, repeat int, interval int) {
	for i := 0; repeat <= 0 || i < repeat; i++ {
		_, err := client.WriteSingleCoil(address, value)
		if err != nil {
			log.Printf("Error during write operation: %v", err)
		} else {
			log.Printf("Successfully wrote single coil: %v", value)
		}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

// writeSingleRegister writes a single register to the Modbus server
func writeSingleRegister(client modbus.Client, address uint16, value uint16, repeat int, interval int) {
	for i := 0; repeat <= 0 || i < repeat; i++ {
		_, err := client.WriteSingleRegister(address, value)
		if err != nil {
			log.Printf("Error during write operation: %v", err)
		} else {
			log.Printf("Successfully wrote single register: %v", value)
		}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

// writeMultipleCoils writes multiple coils to the Modbus server
func writeMultipleCoils(client modbus.Client, start uint16, values []uint16, repeat int, interval int) {
	data := make([]byte, len(values)*2)
	for i, value := range values {
		binary.BigEndian.PutUint16(data[i*2:i*2+2], value)
	}

	for i := 0; repeat <= 0 || i < repeat; i++ {
		_, err := client.WriteMultipleCoils(start, uint16(len(values)), data)
		if err != nil {
			log.Printf("Error during write operation: %v", err)
		} else {
			log.Printf("Successfully wrote multiple coils: %v", values)
		}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

// writeMultipleRegisters writes multiple registers to the Modbus server
func writeMultipleRegisters(client modbus.Client, start uint16, values []uint16, repeat int, interval int) {
	data := make([]byte, len(values)*2)
	for i, value := range values {
		binary.BigEndian.PutUint16(data[i*2:i*2+2], value)
	}

	for i := 0; repeat <= 0 || i < repeat; i++ {
		_, err := client.WriteMultipleRegisters(start, uint16(len(values)), data)
		if err != nil {
			log.Printf("Error during write operation: %v", err)
		} else {
			log.Printf("Successfully wrote multiple registers: %v", values)
		}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}
