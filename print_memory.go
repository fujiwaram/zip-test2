package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

func printMemoryStatsHeader() {
	header := []string{"#", "Alloc", "HeapAlloc", "TotalAlloc", "HeapObjects", "Sys", "NumGC"}
	fmt.Println(strings.Join(header, ","))
}

func printMemoryStats(prefix string) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	data := []string{prefix, toKb(ms.Alloc), toKb(ms.HeapAlloc), toKb(ms.TotalAlloc), toKb(ms.HeapObjects), toKb(ms.Sys), strconv.Itoa(int(ms.NumGC))}
	fmt.Println(strings.Join(data, ","))
}

func toKb(bytes uint64) string {
	return strconv.FormatUint(bytes/1024, 10)
}
