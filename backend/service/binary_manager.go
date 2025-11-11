package service

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gogf/gf/v2/frame/g"
)

// BinaryManager manages binaries lifecycle (embedded or local)
type BinaryManager struct {
	useEmbedded bool
	binaries    embed.FS
	binPath     string
	processes   []*exec.Cmd
	execOrder   []string
	cacheDir    string
}

// BinariesConfig interface for config dependency
type BinariesConfig interface {
	IsEnabled() bool
	IsUseEmbedded() bool
	GetBinPath() string
	GetStartupOrder() []string
}

// NewBinaryManagerFromConfig creates a binary manager from config
func NewBinaryManagerFromConfig(cfg BinariesConfig, binaries embed.FS) (*BinaryManager, error) {
	if cfg == nil || !cfg.IsEnabled() {
		return nil, fmt.Errorf("binary manager is disabled")
	}

	startupOrder := cfg.GetStartupOrder()
	if len(startupOrder) == 0 {
		return nil, fmt.Errorf("no binaries in startup_order")
	}

	binPath := cfg.GetBinPath()
	if binPath == "" {
		binPath = "./bin"
	}

	useEmbedded := cfg.IsUseEmbedded()

	bm, err := NewBinaryManager(useEmbedded, binaries, binPath, startupOrder)
	if err != nil {
		return nil, err
	}

	// Log initialization mode
	ctx := context.Background()
	if useEmbedded {
		g.Log().Info(ctx, "Binary manager: using embedded mode")
	} else {
		g.Log().Infof(ctx, "Binary manager: using local mode (path: %s)", binPath)
	}

	return bm, nil
}

// NewBinaryManager creates a new binary manager
// If useEmbedded is true, binaries will be extracted from embed.FS
// If useEmbedded is false, binaries will be loaded from binPath directory
func NewBinaryManager(useEmbedded bool, binaries embed.FS, binPath string, execOrder []string) (*BinaryManager, error) {
	var cacheDir string

	if useEmbedded {
		// Get app cache directory for embedded mode
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get cache directory: %w", err)
		}

		// Create app-specific cache directory
		cacheDir = filepath.Join(userCacheDir, "wachat", "bin")
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create app cache directory: %w", err)
		}
	} else {
		// Use local bin directory
		if !filepath.IsAbs(binPath) {
			// Convert relative path to absolute
			absPath, err := filepath.Abs(binPath)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve bin path: %w", err)
			}
			binPath = absPath
		}
		cacheDir = binPath
		g.Log().Infof(context.Background(), "Using local bin directory: %s", binPath)
	}

	return &BinaryManager{
		useEmbedded: useEmbedded,
		binaries:    binaries,
		binPath:     binPath,
		processes:   make([]*exec.Cmd, 0),
		execOrder:   execOrder,
		cacheDir:    cacheDir,
	}, nil
}

// StartAll starts all binaries in the specified order
func (bm *BinaryManager) StartAll(ctx context.Context) error {
	successCount := 0
	for _, binaryName := range bm.execOrder {
		if err := bm.startBinary(ctx, binaryName); err != nil {
			g.Log().Warningf(ctx, "Failed to start %s: %v", binaryName, err)
			// Continue with next binary instead of stopping
			continue
		}
		successCount++
	}

	if successCount == 0 {
		return fmt.Errorf("failed to start any binaries")
	}

	g.Log().Infof(ctx, "Started %d/%d binaries successfully", successCount, len(bm.execOrder))
	return nil
}

// startBinary extracts (if embedded) and starts a single binary
func (bm *BinaryManager) startBinary(ctx context.Context, name string) error {
	var executablePath string

	if bm.useEmbedded {
		// Embedded mode: extract from embed.FS
		binaryPath := filepath.Join("bin", name)
		data, err := fs.ReadFile(bm.binaries, binaryPath)
		if err != nil {
			return fmt.Errorf("failed to read embedded binary %s: %w", name, err)
		}

		// Extract to cache directory
		executablePath = filepath.Join(bm.cacheDir, name)
		if err := os.WriteFile(executablePath, data, 0755); err != nil {
			return fmt.Errorf("failed to write binary %s: %w", name, err)
		}
		g.Log().Infof(ctx, "Extracted %s to %s", name, executablePath)
	} else {
		// Local mode: use binary from local directory
		executablePath = filepath.Join(bm.cacheDir, name)

		// Check if binary exists
		if _, err := os.Stat(executablePath); err != nil {
			return fmt.Errorf("binary %s not found at %s: %w", name, executablePath, err)
		}

		// Ensure executable permission
		if err := os.Chmod(executablePath, 0755); err != nil {
			g.Log().Warningf(ctx, "Warning: failed to chmod %s: %v", name, err)
		}

		g.Log().Infof(ctx, "Using local binary: %s", executablePath)
	}

	// Run binary in background
	cmd := exec.CommandContext(ctx, executablePath)
	cmd.Dir = bm.cacheDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", name, err)
	}

	g.Log().Infof(ctx, "%s started successfully (PID: %d)", name, cmd.Process.Pid)

	// Save process reference
	bm.processes = append(bm.processes, cmd)

	// Wait for process in a goroutine
	go func(processName string, process *exec.Cmd) {
		if err := process.Wait(); err != nil {
			g.Log().Warningf(context.Background(), "%s exited with error: %v", processName, err)
		} else {
			g.Log().Infof(context.Background(), "%s exited successfully", processName)
		}
	}(name, cmd)

	return nil
}

// GetProcessCount returns the number of running processes
func (bm *BinaryManager) GetProcessCount() int {
	return len(bm.processes)
}

// Cleanup terminates all managed processes
func (bm *BinaryManager) Cleanup() {
	ctx := context.Background()
	for i, cmd := range bm.processes {
		if cmd.Process != nil {
			g.Log().Infof(ctx, "Terminating process %d (PID: %d)", i, cmd.Process.Pid)
			if err := cmd.Process.Kill(); err != nil {
				g.Log().Warningf(ctx, "Failed to kill process %d: %v", i, err)
			}
		}
	}
}
