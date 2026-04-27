package config

import (
	"bufio"
	"flag"
	"os"
	"strings"
)

type AppConfig struct {
	Debug          bool
	DebugImagesDir string
}

func Load() AppConfig {
	loadedEnv := loadDotEnv(".env")
	if !loadedEnv {
		loadDotEnv(".env.example")
	}

	debugFlag := flag.Bool("debug", false, "Enable debug image output")
	debugDirFlag := flag.String("debug-dir", "", "Directory used for debug image output")
	flag.Parse()

	cfg := AppConfig{
		Debug:          parseBool(os.Getenv("DEBUG")),
		DebugImagesDir: "debug_images",
	}

	if envDir := strings.TrimSpace(os.Getenv("DEBUG_IMAGES_DIR")); envDir != "" {
		cfg.DebugImagesDir = envDir
	}

	if *debugFlag {
		cfg.Debug = true
	}
	if flagDir := strings.TrimSpace(*debugDirFlag); flagDir != "" {
		cfg.DebugImagesDir = flagDir
	}

	return cfg
}

func parseBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "t", "true", "y", "yes", "on":
		return true
	default:
		return false
	}
}

func loadDotEnv(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	loadedAny := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}

		if _, present := os.LookupEnv(key); present {
			continue
		}

		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		_ = os.Setenv(key, value)
		loadedAny = true
	}

	return loadedAny
}
