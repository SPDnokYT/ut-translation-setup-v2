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
	wails.LogInfo(s.ctx, "PickTargetService Iniciado")
}

func (s *PickTargetService) QuickFind() FileValidationResult {
	wails.LogInfo(s.ctx, "Iniciando busca automática pelo jogo (QuickFind)...")
	steamPath := s.findSteamPath()

	for _, gameID := range []int{1574820, 2296400} {
		wails.LogInfo(s.ctx, fmt.Sprintf("Procurando diretório de instalação para o GameID: %d", gameID))
		installDir := s.findGamePath(steamPath, gameID)
		if installDir == "" {
			continue
		}

		candidate := filepath.Join(installDir, "UntilThen.pck")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			wails.LogInfo(s.ctx, fmt.Sprintf("QuickFind: Arquivo encontrado com sucesso em: %s", candidate))
			return s.validateFile(candidate)
		}
	}

	wails.LogWarning(s.ctx, "QuickFind: Arquivo UntilThen.pck não encontrado nos diretórios do Steam.")
	return FileValidationResult{Valid: false, Error: "not_found"}
}

func (s *PickTargetService) OpenFilePicker() FileValidationResult {
	wails.LogInfo(s.ctx, "Abrindo seletor manual de arquivos para o usuário...")

	path, err := wails.OpenFileDialog(s.ctx, wails.OpenDialogOptions{
		Title: "Selecione UntilThen.pck",
		Filters: []wails.FileFilter{
			{DisplayName: "UntilThen.pck (*.pck)", Pattern: "*.pck"},
		},
	})

	if err != nil {
		wails.LogError(s.ctx, fmt.Sprintf("Erro ao abrir seletor de arquivos: %v", err))
		return FileValidationResult{Valid: false, Error: "not_selected"}
	}

	if path == "" {
		wails.LogInfo(s.ctx, "Seleção manual de arquivos cancelada pelo usuário.")
		return FileValidationResult{Valid: false, Error: "not_selected"}
	}

	wails.LogInfo(s.ctx, fmt.Sprintf("Usuário selecionou o arquivo manualmente: %s", path))
	return s.validateFile(path)
}

func (s *PickTargetService) CheckFreeSpace(path string) bool {
	if path == "" {
		return false
	}

	dir := filepath.Dir(path)
	usage, err := disk.Usage(dir)
	if err != nil {
		wails.LogError(s.ctx, fmt.Sprintf("Erro ao verificar espaço em disco no diretório %s: %v", dir, err))
		return false
	}

	const requiredBytes uint64 = 5368709120
	hasSpace := usage.Free >= requiredBytes

	if !hasSpace {
		wails.LogWarning(s.ctx, fmt.Sprintf("Espaço insuficiente. Requerido: %d, Disponível: %d em %s", requiredBytes, usage.Free, dir))
	} else {
		wails.LogInfo(s.ctx, fmt.Sprintf("Espaço em disco validado com sucesso em: %s", dir))
	}

	return hasSpace
}

func (s *PickTargetService) SaveSettings(path string, isDemo bool, makeBackup bool) {
	wails.LogInfo(s.ctx, fmt.Sprintf("Salvando estado da instalação [Path: %s, Demo: %t, Backup: %t]", path, isDemo, makeBackup))
	s.state.SetState(path, isDemo, makeBackup)
}

func (s *PickTargetService) validateFile(path string) FileValidationResult {
	if path == "" {
		return FileValidationResult{Valid: false, Error: "invalid_file"}
	}

	info, err := os.Stat(path)
	if err != nil || info.IsDir() || strings.ToLower(filepath.Ext(path)) != ".pck" {
		wails.LogWarning(s.ctx, fmt.Sprintf("Falha na validação do arquivo (não existe, é diretório ou extensão errada): %s", path))
		return FileValidationResult{Valid: false, Error: "invalid_file"}
	}

	isDemo := strings.Contains(filepath.Dir(path), "Until Then Demo")
	wails.LogInfo(s.ctx, fmt.Sprintf("Arquivo validado com sucesso. Identificado como Demo: %t", isDemo))

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
		wails.LogError(s.ctx, fmt.Sprintf("Falha ao abrir libraryfolders.vdf: %v", err))
		return ""
	}

	data, err := vdf.NewParser(file).Parse()
	file.Close()
	if err != nil {
		wails.LogError(s.ctx, fmt.Sprintf("Falha ao parsear libraryfolders.vdf: %v", err))
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
			wails.LogWarning(s.ctx, fmt.Sprintf("Falha ao parsear %s: %v", manifestPath, err))
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
			wails.LogInfo(s.ctx, fmt.Sprintf("Diretório do jogo encontrado: %s", absolutePath))
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
			wails.LogInfo(s.ctx, fmt.Sprintf("Diretório base do Steam encontrado em: %s", p))
			return p
		}
	}

	return ""
}
