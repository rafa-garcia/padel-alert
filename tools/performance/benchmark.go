package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/config"
	"github.com/rafa-garcia/padel-alert/internal/domain/model"
	"github.com/rafa-garcia/padel-alert/internal/logger"
	"github.com/rafa-garcia/padel-alert/internal/scheduler"
	"github.com/rafa-garcia/padel-alert/internal/storage"
)

var (
	cpuprofile    = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile    = flag.String("memprofile", "", "write memory profile to file")
	testDuration  = flag.Int("duration", 30, "benchmark duration in seconds")
	redisURL      = flag.String("redis", "redis://localhost:6379", "Redis connection URL")
	logLevel      = flag.String("log", "error", "Log level (info, debug, error)")
	benchmarkType = flag.String("type", "all", "Benchmark type: storage, scheduler, api, or all")
)

func main() {
	flag.Parse()

	logger.Init(*logLevel)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("warning: could not close CPU profile file: %v", err)
			}
		}()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("could not load configuration: ", err)
	}

	if *redisURL != "" {
		cfg.RedisURL = *redisURL
	}

	redisClient, err := storage.NewRedisClient(cfg)
	if err != nil {
		log.Fatal("could not connect to Redis: ", err)
	}

	ruleStorage := storage.NewRedisRuleStorage(redisClient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*testDuration)*time.Second)
	defer cancel()

	switch *benchmarkType {
	case "storage":
		benchmarkStorage(ctx, ruleStorage)
	case "scheduler":
		benchmarkScheduler(ctx, cfg, ruleStorage)
	case "api":
		benchmarkAPI(ctx, cfg, ruleStorage)
	case "all":
		benchmarkAll(ctx, cfg, ruleStorage)
	default:
		log.Fatalf("Unknown benchmark type: %s", *benchmarkType)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("warning: could not close memory profile file: %v", err)
			}
		}()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

func benchmarkStorage(ctx context.Context, ruleStorage *storage.RedisRuleStorage) {
	fmt.Println("Running storage benchmark...")

	startTime := time.Now()
	ruleCount := 1000

	for i := 0; i < ruleCount; i++ {
		ruleID := fmt.Sprintf("benchmark-rule-%d", i)

		err := ruleStorage.CreateRule(ctx, &model.Rule{
			ID:      ruleID,
			UserID:  "benchmark-user",
			Email:   "benchmark@example.com",
			Type:    "match",
			Name:    fmt.Sprintf("Benchmark Rule %d", i),
			ClubIDs: []string{"club-1"},
			Active:  true,
		})

		if err != nil {
			log.Printf("Error creating rule: %v", err)
			continue
		}

		err = ruleStorage.ScheduleRule(ctx, ruleID, time.Now().Add(time.Duration(i)*time.Second))
		if err != nil {
			log.Printf("Error scheduling rule: %v", err)
		}
	}

	createDuration := time.Since(startTime)
	fmt.Printf("Created %d rules in %v (%.2f rules/sec)\n",
		ruleCount, createDuration, float64(ruleCount)/createDuration.Seconds())

	startTime = time.Now()
	readCount := 100

	for i := 0; i < readCount; i++ {
		_, err := ruleStorage.GetScheduledRules(ctx, time.Now().Add(time.Duration(i*10)*time.Second))
		if err != nil {
			log.Printf("Error getting scheduled rules: %v", err)
		}
	}

	readDuration := time.Since(startTime)
	fmt.Printf("Read scheduled rules %d times in %v (%.2f reads/sec)\n",
		readCount, readDuration, float64(readCount)/readDuration.Seconds())

	for i := 0; i < ruleCount; i++ {
		ruleID := fmt.Sprintf("benchmark-rule-%d", i)
		err := ruleStorage.DeleteRule(ctx, ruleID)
		if err != nil {
			log.Printf("Error deleting rule: %v", err)
		}
	}
}

func benchmarkScheduler(ctx context.Context, cfg *config.Config, ruleStorage *storage.RedisRuleStorage) {
	fmt.Println("Running scheduler benchmark...")

	sched := scheduler.NewScheduler(cfg, ruleStorage)

	err := sched.Start()
	if err != nil {
		log.Fatalf("Error starting scheduler: %v", err)
	}

	ruleCount := 100
	for i := 0; i < ruleCount; i++ {
		ruleID := fmt.Sprintf("scheduler-rule-%d", i)

		err := ruleStorage.CreateRule(ctx, &model.Rule{
			ID:      ruleID,
			UserID:  "benchmark-user",
			Email:   "benchmark@example.com",
			Type:    "match",
			Name:    fmt.Sprintf("Scheduler Rule %d", i),
			ClubIDs: []string{"club-1"},
			Active:  true,
		})

		if err != nil {
			log.Printf("Error creating rule: %v", err)
			continue
		}

		err = ruleStorage.ScheduleRule(ctx, ruleID, time.Now())
		if err != nil {
			log.Printf("Error scheduling rule: %v", err)
		}
	}

	schedulerRunTime := 10 * time.Second
	fmt.Printf("Scheduler running for %v, processing %d rules...\n", schedulerRunTime, ruleCount)
	time.Sleep(schedulerRunTime)

	sched.Stop()

	for i := 0; i < ruleCount; i++ {
		ruleID := fmt.Sprintf("scheduler-rule-%d", i)
		err := ruleStorage.DeleteRule(ctx, ruleID)
		if err != nil {
			log.Printf("Error deleting rule: %v", err)
		}
	}
}

func benchmarkAPI(_ context.Context, _ *config.Config, _ *storage.RedisRuleStorage) {
	fmt.Println("Running API benchmark...")

	fmt.Println("API router creation benchmark skipped.")

	fmt.Println("API router creation benchmark complete.")
	fmt.Println("For full API benchmarking, consider using a tool like wrk or ab.")
}

func benchmarkAll(ctx context.Context, cfg *config.Config, ruleStorage *storage.RedisRuleStorage) {
	benchmarkStorage(ctx, ruleStorage)
	benchmarkScheduler(ctx, cfg, ruleStorage)
	benchmarkAPI(ctx, cfg, ruleStorage)
}
