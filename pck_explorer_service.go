package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed assets/translation_files.zip
var translationFilesZip []byte

type PckExplorerService struct {
	ctx   context.Context
	state *InstallerState
}

func NewPckExplorerService(state *InstallerState) *PckExplorerService {
	return &PckExplorerService{state: state}
}

func (s *PckExplorerService) startup(ctx context.Context) {
	s.ctx = ctx
	wails.LogInfo(s.ctx, "Служба PckExplorerService запущена")
}

func (s *PckExplorerService) RunInstallation() {
	wails.LogInfo(s.ctx, "Получен запрос на установку. Запуск фонового процесса...")
	go s.startInstallProcess()
}

func (s *PckExplorerService) startInstallProcess() {
	targetPckPath, isDemo, makeBackup := s.state.GetState()
	wails.LogInfo(s.ctx, fmt.Sprintf("Параметры установки - Путь: %s, Демо: %t, Резервная копия: %t", targetPckPath, isDemo, makeBackup))

	gameDir := filepath.Dir(targetPckPath)
	modifiedPckPath := filepath.Join(gameDir, "ModifiedPCK.pck")
	backupPckPath := filepath.Join(gameDir, "UntilThen.pck.bak")

	tempDir, err := os.MkdirTemp("", "untilthen_patcher_*")
	if err != nil {
		s.failAndLog(modifiedPckPath, fmt.Errorf("не удалось создать временную папку: %w", err))
		return
	}
	// Garante a limpeza do diretório temporário no final do processo
	defer func() {
		wails.LogInfo(s.ctx, fmt.Sprintf("Очистка временного каталога: %s", tempDir))
		os.RemoveAll(tempDir)
	}()

	wails.LogInfo(s.ctx, fmt.Sprintf("Временный каталог успешно создан в: %s", tempDir))

	wails.EventsEmit(s.ctx, "install_step", "Извлечение инструментов для патча...")
	wails.LogInfo(s.ctx, "Начало извлечения бинарного файла pckExplorerBinZip...")
	if err := s.unzipFromMemory(pckExplorerBinZip, tempDir, "unzip_bin_progress"); err != nil {
		s.failAndLog(modifiedPckPath, fmt.Errorf("ошибка при извлечении инструментов: %w", err))
		return
	}

	binPath := filepath.Join(tempDir, pckBinName)
	wails.LogInfo(s.ctx, fmt.Sprintf("Задан путь к бинарному файлу: %s", binPath))

	if runtime.GOOS != "windows" {
		wails.LogInfo(s.ctx, "Обнаружена система, отличная от Windows, применение прав на выполнение (0755) к бинарному файлу.")
		os.Chmod(binPath, 0755)
	}

	wails.EventsEmit(s.ctx, "install_step", "Подготовка файлов перевода...")
	wails.LogInfo(s.ctx, "Начало извлечения файлов перевода (translationFilesZip)...")
	if err := s.unzipFromMemory(translationFilesZip, tempDir, "unzip_trans_progress"); err != nil {
		s.failAndLog(modifiedPckPath, fmt.Errorf("ошибка при извлечении файлов перевода: %w", err))
		return
	}

	translationFolder := "full"
	if isDemo {
		translationFolder = "demo"
	}
	translationFilesPath := filepath.Join(tempDir, translationFolder)
	wails.LogInfo(s.ctx, fmt.Sprintf("Выбранный режим перевода: %s (Путь: %s)", translationFolder, translationFilesPath))

	wails.EventsEmit(s.ctx, "install_step", "Установка перевода (это может занять несколько минут)...")

	wails.LogInfo(s.ctx, fmt.Sprintf("Запуск команды: %s -pc %s %s %s 2.2.4.1", binPath, targetPckPath, translationFilesPath, modifiedPckPath))
	cmd := exec.CommandContext(s.ctx, binPath, "-pc", targetPckPath, translationFilesPath, modifiedPckPath, "2.2.4.1")

	// This prevents the console window from appearing on Windows
	cmd.SysProcAttr = getSysProcAttr()

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		s.failAndLog(modifiedPckPath, fmt.Errorf("не удалось запустить процесс установки патча: %w", err))
		return
	}

	wails.LogInfo(s.ctx, "Процесс установки патча запущен. Ожидание завершения...")
	go s.streamLogs(stdout, "install_log")
	go s.streamLogs(stderr, "install_error")

	if err := cmd.Wait(); err != nil {
		s.failAndLog(modifiedPckPath, fmt.Errorf("ошибка при выполнении патчера: %w", err))
		return
	}

	wails.LogInfo(s.ctx, "Процесс установки патча успешно завершен.")
	wails.EventsEmit(s.ctx, "install_step", "Завершение установки...")

	if makeBackup {
		wails.LogInfo(s.ctx, fmt.Sprintf("Создание резервной копии оригинального PCK в: %s", backupPckPath))
		os.Remove(backupPckPath) // Remove existing if any
		if err := os.Rename(targetPckPath, backupPckPath); err != nil {
			s.failAndLog(modifiedPckPath, fmt.Errorf("ошибка при создании резервной копии: %w", err))
			return
		}
	} else {
		wails.LogInfo(s.ctx, "Резервное копирование отключено. Удаление оригинального файла PCK...")
		os.Remove(targetPckPath)
	}

	wails.LogInfo(s.ctx, fmt.Sprintf("Переименование измененного PCK с %s на %s", modifiedPckPath, targetPckPath))
	if err := os.Rename(modifiedPckPath, targetPckPath); err != nil {
		wails.LogError(s.ctx, fmt.Sprintf("Не удалось переименовать финальный PCK: %v", err))
		if makeBackup {
			wails.LogInfo(s.ctx, "Попытка восстановления из резервной копии из-за ошибки переименования...")
			os.Rename(backupPckPath, targetPckPath)
		}
		s.failAndLog(modifiedPckPath, fmt.Errorf("ошибка при переименовании финального файла: %w", err))
		return
	}

	wails.LogInfo(s.ctx, "Установка успешно завершена!")
	wails.EventsEmit(s.ctx, "install_success", "Успешно!")
}

func (s *PckExplorerService) unzipFromMemory(data []byte, dest, eventName string) error {
	wails.LogInfo(s.ctx, fmt.Sprintf("Распаковка %d байт в %s...", len(data), dest))
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	total := len(reader.File)
	for i, f := range reader.File {
		progress := int(float64(i+1) / float64(total) * 100)
		wails.EventsEmit(s.ctx, eventName, progress)

		normalizedName := strings.ReplaceAll(f.Name, "\\", "/")
		fpath := filepath.Join(dest, normalizedName)

		isDir := f.FileInfo().IsDir() || strings.HasSuffix(normalizedName, "/")

		if isDir {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				wails.LogError(s.ctx, fmt.Sprintf("Ошибка при создании структуры каталогов для %s: %v", fpath, err))
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			wails.LogError(s.ctx, fmt.Sprintf("Ошибка при создании родительского каталога для файла %s: %v", fpath, err))
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			wails.LogError(s.ctx, fmt.Sprintf("Ошибка при подготовке файла %s для записи: %v", fpath, err))
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			wails.LogError(s.ctx, fmt.Sprintf("Ошибка при чтении файла из архива %s: %v", f.Name, err))
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			wails.LogError(s.ctx, fmt.Sprintf("Ошибка при копировании извлеченных данных в %s: %v", fpath, err))
			return err
		}
	}

	wails.LogInfo(s.ctx, fmt.Sprintf("Распаковка %d файлов завершена.", total))
	return nil
}

func (s *PckExplorerService) failAndLog(modifiedPck string, err error) {
	wails.LogError(s.ctx, fmt.Sprintf("Критическая ошибка при установке: %v", err))
	wails.EventsEmit(s.ctx, "install_error", err.Error())

	if modifiedPck != "" {
		wails.LogInfo(s.ctx, fmt.Sprintf("Удаление частично созданного ModifiedPCK из-за ошибки: %s", modifiedPck))
		os.Remove(modifiedPck)
	}

	wails.MessageDialog(s.ctx, wails.MessageDialogOptions{
		Type:    wails.ErrorDialog,
		Title:   "Ошибка во время установки",
		Message: fmt.Sprintf("Во время установки произошла непредвиденная ошибка\nФайл лога создан в: %s", GetLogFilePath()),
	})
}

func (s *PckExplorerService) streamLogs(pipe io.ReadCloser, eventName string) {
	defer pipe.Close()
	scanner := bufio.NewScanner(pipe)

	var lastLine string
	hasNewContent := false

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	go func() {
		for scanner.Scan() {
			lastLine = scanner.Text()
			hasNewContent = true
		}
	}()

	for {
		select {
		case <-ticker.C:
			if hasNewContent {
				wails.EventsEmit(s.ctx, eventName, lastLine)
				hasNewContent = false
			}
		case <-s.ctx.Done():
			return
		}
	}
}
