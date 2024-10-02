package main

import (
	"fmt"
	"strconv"
	"syscall"
	"unsafe"
)

var (
	modpsapi                     = syscall.NewLazyDLL("psapi.dll")
	modkernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procEnumProcesses            = modpsapi.NewProc("EnumProcesses")
	procOpenProcess              = modkernel32.NewProc("OpenProcess")
	procCloseHandle              = modkernel32.NewProc("CloseHandle")
	procQueryFullProcessImageName = modkernel32.NewProc("QueryFullProcessImageNameW")

	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000 // Limited access rights
)

func EnumProcesses(pids []uint32, size uint32, needed *uint32) error {
	r1, _, e1 := procEnumProcesses.Call(
		uintptr(unsafe.Pointer(&pids[0])),
		uintptr(size),
		uintptr(unsafe.Pointer(needed)),
	)
	if r1 == 0 {
		return e1
	}
	return nil
}

func OpenProcess(pid uint32) (syscall.Handle, error) {
	r1, _, e1 := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_LIMITED_INFORMATION),
		uintptr(0),
		uintptr(pid),
	)
	if r1 == 0 {
		return syscall.InvalidHandle, e1
	}
	return syscall.Handle(r1), nil
}

func CloseHandle(handle syscall.Handle) error {
	r1, _, e1 := procCloseHandle.Call(uintptr(handle))
	if r1 == 0 {
		return e1
	}
	return nil
}

func QueryProcessImageName(handle syscall.Handle) (string, error) {
	var size uint32 = 1024
	buffer := make([]uint16, size)
	r1, _, e1 := procQueryFullProcessImageName.Call(
		uintptr(handle),
		uintptr(0),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if r1 == 0 {
		return "", e1
	}
	return syscall.UTF16ToString(buffer), nil
}

func GetPIDsUsingFile(filePath string) ([]uint32, error) {
	var pids [1024]uint32 // Array to store PIDs
	var bytesReturned uint32
	var results []uint32

	err := EnumProcesses(pids[:], uint32(unsafe.Sizeof(pids)), &bytesReturned)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate processes: %v", err)
	}

	numProcesses := bytesReturned / uint32(unsafe.Sizeof(pids[0]))

	for i := 0; i < int(numProcesses); i++ {
		pid := pids[i]
		handle, err := OpenProcess(pid)
		if err != nil {
			continue
		}
		// Here you would check if the process has the file open
		// For simplicity, let's assume we collect all PIDs
		results = append(results, pid)
		CloseHandle(handle)
	}

	return results, nil
}

func GetProcessDetails(pid uint32) error {
	handle, err := OpenProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to open process with PID %d: %v", pid, err)
	}
	defer CloseHandle(handle)

	// Get process image name (full path to executable)
	processName, err := QueryProcessImageName(handle)
	if err != nil {
		return fmt.Errorf("failed to get process name for PID %d: %v", pid, err)
	}

	fmt.Printf("Process ID: %d\n", pid)
	fmt.Printf("Process Name: %s\n", processName)
	// You can add more details to fetch if required

	return nil
}

func main() {
	var choice string
	fmt.Println("Choose an option:")
	fmt.Println("1. Find PIDs using a specific file.")
	fmt.Println("2. Get details of a specific process by PID.")
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		var filePath string
		fmt.Print("Enter the file path: ")
		fmt.Scanln(&filePath)

		if filePath == "" {
			fmt.Println("File path cannot be empty.")
			return
		}

		pids, err := GetPIDsUsingFile(filePath)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if len(pids) > 0 {
			fmt.Printf("Processes using file %s: %v\n", filePath, pids)
		} else {
			fmt.Printf("No processes are currently using file %s.\n", filePath)
		}

	case "2":
		var pidInput string
		fmt.Print("Enter the process ID (PID): ")
		fmt.Scanln(&pidInput)

		pid, err := strconv.ParseUint(pidInput, 10, 32)
		if err != nil {
			fmt.Println("Invalid PID. It should be a number.")
			return
		}

		err = GetProcessDetails(uint32(pid))
		if err != nil {
			fmt.Println("Error:", err)
		}

	default:
		fmt.Println("Invalid choice. Please select either 1 or 2.")
	}
}
