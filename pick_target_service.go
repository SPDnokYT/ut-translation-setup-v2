package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/andygrunwald/vdf"
	"github.com/shirou/gopsutil/disk"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

type FileValidationResult struct {
	Valid  bool   `json:"valid"`
	IsDemo bool   `json:"isDemo"`
	Path   string `json:"path"`
	Error  string `json:"error,omitempty"` // "not_found" | "not_selected" | "invalid_file"
}

type PickTargetService struct {
	ctx   context.Context
	state *InstallerState
}

func NewPickTargetService(state *InstallerState) *PickTargetService {
	return &PickTargetService{
		state: state,
	}
}

func (s *PickTargetService) startup(ctx context.Context) {
	s.ctx = ctx
	wails.LogInfo(s.ctx, "Служба PickTargetService запущена")
}

func (s *PickTargetService) QuickFind() FileValidationResult {
	wails.LogInfo(s.ctx, "Начало автоматического поиска игры (QuickFind)...")
	steamPath := s.findSteamPath()

	for _, gameID := range []int{1574820, 2296400} {
		wails.LogInfo(s.ctx, fmt.Sprintf("Поиск каталога установки для GameID: %d", gameID))
		installDir := s.findGamePath(steamPath, gameID)
		if installDir == "" {
			continue
		}

		candidate := filepath.Join(installDir, "UntilThen.pck")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			wails.LogInfo(s.ctx, fmt.Sprintf("QuickFind: Файл успешно найден в:%s", candidate))
			return s.validateFile(candidate)
		}
	}

	wails.LogWarning(s.ctx, "QuickFind: Файл UntilThen.pck не найден в каталогах Steam.")
	return FileValidationResult{Valid: false, Error: "not_found"}
}

func (s *PickTargetService) OpenFilePicker() FileValidationResult {
	wails.LogInfo(s.ctx, "Открытие ручного выбора файлов для пользователя...")

	path, err := wails.OpenFileDialog(s.ctx, wails.OpenDialogOptions{
		Title: "Выберите UntilThen.pck",
		Filters: []wails.FileFilter{
			{DisplayName: "UntilThen.pck (*.pck)", Pattern: "*.pck"},
		},
	})

	if err != nil {
		wails.LogError(s.ctx, fmt.Sprintf("Ошибка при открытии окна выбора файлов: %v", err))
		return FileValidationResult{Valid: false, Error: "not_selected"}
	}

	if path == "" {
		wails.LogInfo(s.ctx, "Ручной выбор файлов отменен пользователем.")
		return FileValidationResult{Valid: false, Error: "not_selected"}
	}

	wails.LogInfo(s.ctx, fmt.Sprintf("Пользователь выбрал файл вручную: %s", path))
	return s.validateFile(path)
}

func (s *PickTargetService) CheckFreeSpace(path string) bool {
	if path == "" {
		return false
	}

	dir := filepath.Dir(path)
	usage, err := disk.Usage(dir)
	if err != nil {
		wails.LogError(s.ctx, fmt.Sprintf("Ошибка при проверке места на диске в каталоге %s: %v", dir, err))
		return false
	}

	const requiredBytes uint64 = 5368709120
	hasSpace := usage.Free >= requiredBytes

	if !hasSpace {
		wails.LogWarning(s.ctx, fmt.Sprintf("Недостаточно места. Требуется: %d, Доступно: %d в %s", requiredBytes, usage.Free, dir))
	} else {
		wails.LogInfo(s.ctx, fmt.Sprintf("Проверка места на диске успешно завершена в: %s", dir))
	}

	return hasSpace
}

func (s *PickTargetService) SaveSettings(path string, isDemo bool, makeBackup bool) {
	wails.LogInfo(s.ctx, fmt.Sprintf("Сохранение состояния установки [Путь: %s, Демо: %t, Резервная копия: %t]", path, isDemo, makeBackup))
	s.state.SetState(path, isDemo, makeBackup)
}

func (s *PickTargetService) validateFile(path string) FileValidationResult {
	if path == "" {
		return FileValidationResult{Valid: false, Error: "invalid_file"}
	}

	info, err := os.Stat(path)
	if err != nil || info.IsDir() || strings.ToLower(filepath.Ext(path)) != ".pck" {
		wails.LogWarning(s.ctx, fmt.Sprintf("Ошибка проверки файла (не существует, является каталогом или имеет неверное расширение): %s", path))
		return FileValidationResult{Valid: false, Error: "invalid_file"}
	}

	isDemo := strings.Contains(filepath.Dir(path), "Until Then Demo")
	wails.LogInfo(s.ctx, fmt.Sprintf("Файл успешно проверен. Определен как Demo: %t", isDemo))

	return FileValidationResult{
		Valid:  true,
		IsDemo: isDemo,
		Path:   path,
	}
}

func (s *PickTargetService) findGamePath(steamPath string, gameID int) string {
	if steamPath == "" {
		return ""
	}

	libraryFoldersPath := filepath.Join(steamPath, "steamapps", "libraryfolders.vdf")
	file, err := os.Open(libraryFoldersPath)
	if err != nil {
		wails.LogError(s.ctx, fmt.Sprintf("Не удалось открыть libraryfolders.vdf: %v", err))
		return ""
	}

	data, err := vdf.NewParser(file).Parse()
	file.Close()
	if err != nil {
		wails.LogError(s.ctx, fmt.Sprintf("Не удалось обработать libraryfolders.vdf: %v", err))
		return ""
	}

	folders, ok := data["libraryfolders"].(map[string]any)
	if !ok {
		return ""
	}

	var libraries []string
	for _, entry := range folders {
		entryMap, ok := entry.(map[string]any)
		if !ok {
			continue
		}

		if path, ok := entryMap["path"].(string); ok && path != "" {
			libraries = append(libraries, path)
		}
	}
	if len(libraries) == 0 {
		libraries = []string{steamPath}
	}

	for _, lib := range libraries {
		manifestPath := filepath.Join(lib, "steamapps", fmt.Sprintf("appmanifest_%d.acf", gameID))

		manifestFile, err := os.Open(manifestPath)
		if err != nil {
			continue
		}

		manifest, err := vdf.NewParser(manifestFile).Parse()
		manifestFile.Close()
		if err != nil {
			wails.LogWarning(s.ctx, fmt.Sprintf("Не удалось обработать %s: %v", manifestPath, err))
			continue
		}

		appState, ok := manifest["AppState"].(map[string]any)
		if !ok {
			continue
		}

		installDir, ok := appState["installdir"].(string)
		if !ok || installDir == "" {
			continue
		}

		gamePath := filepath.Join(lib, "steamapps", "common", installDir)
		if info, err := os.Stat(gamePath); err == nil && info.IsDir() {
			absolutePath, _ := filepath.Abs(gamePath)
			wails.LogInfo(s.ctx, fmt.Sprintf("Каталог игры найден: %s", absolutePath))
			return absolutePath
		}
	}

	return ""
}

func (s *PickTargetService) findSteamPath() string {
	var candidates []string

	switch runtime.GOOS {
	case "linux":
		home, _ := os.UserHomeDir()
		candidates = []string{
			filepath.Join(home, ".steam", "steam"),
			filepath.Join(home, ".local", "share", "Steam"),
			filepath.Join(home, "snap", "steam", "common", ".local", "share", "Steam"),
			filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", "data", "Steam"),
		}
	case "windows":
		if path := steamPathFromRegistry(); path != "" {
			candidates = append(candidates, path)
		}
		candidates = append(candidates,
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Steam"),
			filepath.Join(os.Getenv("ProgramFiles"), "Steam"),
		)
	default:
		return ""
	}

	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			wails.LogInfo(s.ctx, fmt.Sprintf("Базовый каталог Steam найден в: %s", p))
			return p
		}
	}

	return ""
}
