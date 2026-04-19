package main

import (
	"os/exec"
	"strings"

	"github.com/webview/webview_go"
)

type App struct {
	w       webview.WebView
	m       *Manager
	config  *Config
}

func NewApp() *App {
	cfg := LoadConfig()
	m := NewManager()
	m.DesktopIcon = cfg.DesktopIcon
	m.DesktopPath = cfg.DesktopPath
	m.NoExtractIcon = cfg.NoExtractIcon

	return &App{
		m:      m,
		config: cfg,
	}
}

// Bindings for the frontend
func (a *App) GetInstalledApps() []installedEntry {
	return a.m.FindInstalledEntries()
}

func (a *App) SelectAppImage() string {
	// Use zenity to pick a file
	cmd := exec.Command("zenity", "--file-selection", "--title=AppImage auswählen", "--file-filter=AppImage (*.AppImage *.appimage) | *.AppImage *.appimage")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (a *App) Install(path, name, desc, icon, args, cats string) string {
	res, err := a.m.Install(path, name, desc, icon, args, cats)
	if err != nil {
		return "FEHLER: " + err.Error()
	}
	return res
}

func (a *App) Uninstall(name string) string {
	entries := a.m.FindInstalledEntries()
	for _, e := range entries {
		if e.Name == name {
			err := a.m.Uninstall(e)
			if err != nil {
				return "FEHLER: " + err.Error()
			}
			return "OK"
		}
	}
	return "App nicht gefunden"
}

func (a *App) GetConfig() *Config {
	return a.config
}

func (a *App) SaveConfig(cfg Config) string {
	*a.config = cfg
	a.m.DesktopIcon = cfg.DesktopIcon
	a.m.DesktopPath = cfg.DesktopPath
	a.m.NoExtractIcon = cfg.NoExtractIcon
	err := a.config.Save()
	if err != nil {
		return err.Error()
	}
	return "OK"
}

// Helper for opening file dialog for desktop path
func (a *App) SelectFolder() string {
	cmd := exec.Command("zenity", "--file-selection", "--directory", "--title=Ordner auswählen")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (a *App) CheckForUpdates() string {
	return GetUpdateVersion()
}
