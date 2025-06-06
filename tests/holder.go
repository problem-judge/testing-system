package tests

import (
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"github.com/xorcare/pointer"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"testing_system/common"
	"testing_system/common/config"
	"testing_system/common/db/models"
	"testing_system/invoker"
	"testing_system/lib/logger"
	"testing_system/master"
	"testing_system/storage"
)

const defaultLogLevel = logger.LogLevelInfo

type TSHolder struct {
	ts *common.TestingSystem
	t  *testing.T

	dir        string
	storageDir string
	submitsDir string

	client *resty.Client

	finishWait sync.WaitGroup

	submits []*submitTest
}

func initTS(t *testing.T, sandbox string) *TSHolder {
	h := &TSHolder{
		t:   t,
		dir: t.TempDir(),
	}
	h.storageDir = filepath.Join(h.dir, "storage")
	h.copyDir("testdata/storage", h.storageDir)

	h.submitsDir = filepath.Join(h.dir, "submits")
	h.copyDir("testdata/submits", h.submitsDir)

	configDir := filepath.Join(h.dir, "configs")
	h.copyDir("testdata/configs", configDir)

	configPath := filepath.Join(configDir, "config.yaml")
	h.initTSConfig(configPath, sandbox)

	h.ts = common.InitTestingSystem(configPath)

	h.client = resty.New().SetBaseURL("http://localhost:" + strconv.Itoa(h.ts.Config.Port))

	h.addProblems()

	require.NoError(t, invoker.SetupInvoker(h.ts))
	require.NoError(t, master.SetupMaster(h.ts))
	require.NoError(t, storage.SetupStorage(h.ts))

	h.finishWait.Add(1)

	return h
}

func (h *TSHolder) copyDir(src string, dst string) {
	require.NoError(h.t, exec.Command("cp", "-r", src, dst).Run()) // Why go does not have analog???
}

func (h *TSHolder) initTSConfig(configPath string, sandbox string) {
	configContent, err := os.ReadFile(configPath)
	require.NoError(h.t, err)
	cfg := new(config.Config)
	require.NoError(h.t, yaml.Unmarshal(configContent, cfg))
	cfg.Storage.StoragePath = h.storageDir

	cfg.Invoker.SandboxHomePath = filepath.Join(h.dir, "sandbox")
	require.NoError(h.t, os.MkdirAll(cfg.Invoker.SandboxHomePath, 0755))

	cfg.Invoker.CachePath = filepath.Join(h.dir, "invoker_cache")
	require.NoError(h.t, os.MkdirAll(cfg.Invoker.CachePath, 0755))

	cfg.Invoker.CompilerConfigsFolder = filepath.Join(filepath.Dir(configPath), "compiler")

	cfg.Invoker.SandboxType = sandbox
	cfg.Logger = &logger.Config{
		LogLevel: pointer.Int(defaultLogLevel),
	}

	configContent, err = yaml.Marshal(cfg)
	require.NoError(h.t, err)
	require.NoError(h.t, os.WriteFile(configPath, configContent, 0644))
}

func (h *TSHolder) addProblems() {
	h.addProblem(1)
	h.addProblem(2)
	// TODO: Add more
}

func (h *TSHolder) addProblem(id uint) {
	probPath := filepath.Join(h.storageDir, "Problem", strconv.FormatUint(uint64(id), 10))

	probContent, err := os.ReadFile(filepath.Join(probPath, "problem.yaml"))
	require.NoError(h.t, err)

	prob := new(models.Problem)
	require.NoError(h.t, yaml.Unmarshal(probContent, prob))

	prob.ID = id
	require.NoError(h.t, h.ts.DB.Save(prob).Error)

	testlib, err := os.ReadFile(filepath.Join(h.storageDir, "testlib.h"))
	require.NoError(h.t, err)
	require.NoError(h.t, os.WriteFile(filepath.Join(probPath, "sources", "testlib.h"), testlib, 0644))

	cmd := exec.Command("g++", "check.cpp", "-std=c++17", "-o", "check")
	cmd.Dir = filepath.Join(probPath, "sources")
	require.NoError(h.t, cmd.Run())

	checker, err := os.ReadFile(filepath.Join(probPath, "sources", "check"))
	require.NoError(h.t, err)
	require.NoError(h.t, os.MkdirAll(filepath.Join(probPath, "checker"), 0755))
	require.NoError(h.t, os.WriteFile(filepath.Join(probPath, "checker", "check"), checker, 0755))
}

func (h *TSHolder) stop() {
	logger.Warn("Stopping TS because testing is complete")
	h.ts.Stop()
	h.finishWait.Wait()
}

func (h *TSHolder) start() {
	h.ts.Run()
	h.finishWait.Done()
}
