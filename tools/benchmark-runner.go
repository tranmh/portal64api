package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	fmt.Println("Portal64 API Performance Benchmark Suite")
	fmt.Println("========================================")
	
	// Change to project directory
	projectDir := "C:\\Users\\tranm\\work\\svw.info\\portal64api"
	if err := os.Chdir(projectDir); err != nil {
		log.Fatal(err)
	}
	
	// Run benchmark suite
	benchmarks := []struct {
		name    string
		pattern string
		desc    string
	}{
		{"Import System Performance", "BenchmarkSCP.*|BenchmarkZIP.*|BenchmarkDatabase.*", "Tests SCP download, ZIP extraction, and database import performance"},
		{"Status Tracking Performance", "BenchmarkStatusTracker.*", "Tests status tracking and concurrent access performance"},
		{"Service Performance", "BenchmarkImportService.*", "Tests import service status and log retrieval performance"},
		{"Memory Usage", "BenchmarkMemoryUsage.*", "Tests memory usage patterns during import workflow"},
		{"Concurrent Access", "BenchmarkConcurrentAccess.*", "Tests concurrent access patterns and scalability"},
	}
	
	for _, bench := range benchmarks {
		fmt.Printf("\n%s\n", bench.name)
		fmt.Printf("%s\n", bench.desc)
		fmt.Printf("%s\n", strings.Repeat("-", len(bench.name)))
		
		cmd := exec.Command("go", "test", "./tests/benchmarks", "-bench", bench.pattern, "-benchmem", "-count=1", "-timeout=60s")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		start := time.Now()
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Benchmark %s failed: %v\n", bench.name, err)
		}
		duration := time.Since(start)
		fmt.Printf("Completed in: %v\n", duration)
	}
	
	fmt.Println("\n========================================")
	fmt.Println("Performance Analysis Complete")
	fmt.Printf("Total execution time: %v\n", time.Now().Format("15:04:05"))
}
