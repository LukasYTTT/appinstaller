package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/webview/webview_go"
)

//go:embed frontend
var frontendAssets embed.FS

const desktopTemplate = `[Desktop Entry]
Name=%s
Comment=%s
Exec=%s%s
Icon=%s
Type=Application
Categories=%s
StartupNotify=true
StartupWMClass=%s
Terminal=false
`

type Manager struct {
	DryRun         bool
	NoExtractIcon  bool
	DesktopIcon    bool
	DesktopPath    string
	ApplicationsDir string
	IconsDir       string
}

func NewManager() *Manager {
	homeDir, _ := os.UserHomeDir()
	return &Manager{
		DesktopPath:     filepath.Join(homeDir, "Schreibtisch"),
		ApplicationsDir: filepath.Join(homeDir, ".local", "share", "applications"),
		IconsDir:        filepath.Join(homeDir, ".local", "share", "icons", "appimage-install"),
	}
}

// Gültige Freedesktop-Kategorien
// https://specifications.freedesktop.org/menu-spec/latest/category-registry.html
var validCategories = map[string]string{
	"audiovisual":      "AudioVideo",
	"audio":            "Audio",
	"video":            "Video",
	"development":      "Development",
	"education":        "Education",
	"game":             "Game",
	"graphics":         "Graphics",
	"network":          "Network",
	"office":           "Office",
	"science":          "Science",
	"settings":         "Settings",
	"system":           "System",
	"utility":          "Utility",
	"2dgraphics":       "2DGraphics",
	"3dgraphics":       "3DGraphics",
	"accessibility":    "Accessibility",
	"archiving":        "Archiving",
	"astronomy":        "Astronomy",
	"biology":          "Biology",
	"building":         "Building",
	"calculator":       "Calculator",
	"calendar":         "Calendar",
	"chat":             "Chat",
	"chemistry":        "Chemistry",
	"clock":            "Clock",
	"compression":      "Compression",
	"contactmanager":   "ContactManager",
	"database":         "Database",
	"debugger":         "Debugger",
	"dictionary":       "Dictionary",
	"email":            "Email",
	"emulator":         "Emulator",
	"engineering":      "Engineering",
	"filemanager":      "FileManager",
	"filesystem":       "Filesystem",
	"filetransfer":     "FileTransfer",
	"ide":              "IDE",
	"imaging":          "Imaging",
	"instantmessaging": "InstantMessaging",
	"maps":             "Maps",
	"math":             "Math",
	"monitor":          "Monitor",
	"music":            "Music",
	"news":             "News",
	"p2p":              "P2P",
	"photography":      "Photography",
	"player":           "Player",
	"presentation":     "Presentation",
	"printing":         "Printing",
	"recorder":         "Recorder",
	"rss":              "RSS",
	"scanner":          "Scanner",
	"security":         "Security",
	"sequencer":        "Sequencer",
	"spreadsheet":      "Spreadsheet",
	"terminal":         "TerminalEmulator",
	"texteditor":       "TextEditor",
	"texttools":        "TextTools",
	"viewer":           "Viewer",
	"vnc":              "RemoteAccess",
	"webdevelopment":   "WebDevelopment",
	"webbrowser":       "WebBrowser",
	"wordprocessor":    "WordProcessor",
}

func main() {
	// Start silent background update check
	go CheckUpdate()

	// Fix for "Error 71 (Protokollfehler) dispatching to Wayland display"
	// This is a common issue with WebKitGTK on Wayland, especially with NVIDIA drivers.
	if os.Getenv("WEBKIT_DISABLE_DMABUF_RENDERER") == "" {
		os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}

	if len(os.Args) < 2 {
		// Launch GUI if no arguments
		launchGUI()
		return
	}

	installCmd := flag.NewFlagSet("install", flag.ExitOnError)
	uninstallCmd := flag.NewFlagSet("uninstall", flag.ExitOnError)
	guiCmd := flag.NewFlagSet("gui", flag.ExitOnError)

	// install flags
	appImagePath     := installCmd.String("appimage", "", "Pfad zur AppImage-Datei (erforderlich)")
	customName       := installCmd.String("name", "", "Benutzerdefinierter Anwendungsname")
	customDesc       := installCmd.String("description", "", "Kurzbeschreibung der Anwendung")
	customIcon       := installCmd.String("icon", "", "Pfad zu einem benutzerdefinierten Icon (.png/.svg/.xpm)")
	customArgs       := installCmd.String("args", "", "Zusaetzliche Start-Argumente, z.B. \"--no-sandbox\"")
	customCategories := installCmd.String("categories", "", "Kommagetrennte Kategorien, z.B. \"Network,Chat\"")
	listCategories   := installCmd.Bool("list-categories", false, "Alle verfuegbaren Kategorien auflisten und beenden")
	desktop          := installCmd.Bool("desktop", false, "Auch eine .desktop-Datei auf dem Desktop erstellen")
	noExtractIcon    := installCmd.Bool("no-extract-icon", false, "Icon-Extraktion aus dem AppImage ueberspringen")
	dryRun           := installCmd.Bool("dry-run", false, "Vorschau anzeigen, ohne Dateien zu schreiben")

	// uninstall flags
	uninstallName    := uninstallCmd.String("name", "", "Name der zu deinstallierenden Anwendung")
	uninstallList    := uninstallCmd.Bool("list", false, "Alle per appimage-install installierten Apps auflisten")
	uninstallDryRun  := uninstallCmd.Bool("dry-run", false, "Vorschau anzeigen, ohne Dateien zu loeschen")

	if len(os.Args) < 2 {
		globalUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		installCmd.Usage = installUsage
		_ = installCmd.Parse(os.Args[2:])
		if *listCategories {
			printCategories()
			return
		}
		runInstall(appImagePath, customName, customDesc, customIcon, customArgs,
			customCategories, desktop, noExtractIcon, dryRun)

	case "uninstall", "remove":
		uninstallCmd.Usage = uninstallUsage
		_ = uninstallCmd.Parse(os.Args[2:])
		runUninstall(uninstallName, uninstallList, uninstallDryRun)

	case "gui":
		_ = guiCmd.Parse(os.Args[2:])
		launchGUI()

	case "--help", "-h", "help":
		globalUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unbekannter Befehl: %s\n\n", os.Args[1])
		globalUsage()
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// install
// ---------------------------------------------------------------------------
func (m *Manager) Install(appImagePath, customName, customDesc, customIcon, customArgs, customCategories string) (string, error) {
	absAppImage, err := filepath.Abs(appImagePath)
	if err != nil {
		return "", fmt.Errorf("Fehler beim Aufloesen des AppImage-Pfades: %v", err)
	}
	if !fileExists(absAppImage) {
		return "", fmt.Errorf("AppImage nicht gefunden: %s", absAppImage)
	}

	if !m.DryRun {
		if err := os.Chmod(absAppImage, 0755); err != nil {
			return "", fmt.Errorf("Konnte AppImage nicht ausfuehrbar machen: %v", err)
		}
	}

	appName := customName
	if appName == "" {
		base := filepath.Base(absAppImage)
		base = strings.TrimSuffix(base, ".AppImage")
		base = strings.TrimSuffix(base, ".appimage")
		appName = sanitizeName(base)
	}

	description := customDesc
	if description == "" {
		description = "Installed via appimage-install"
	}

	iconPath := ""
	if customIcon != "" {
		absIcon, err := filepath.Abs(customIcon)
		if err != nil || !fileExists(absIcon) {
			return "", fmt.Errorf("Benutzerdefiniertes Icon nicht gefunden: %s", customIcon)
		}
		iconPath = absIcon
	} else if !m.NoExtractIcon {
		iconPath = m.extractIcon(absAppImage, appName)
	}
	if iconPath == "" {
		iconPath = "application-x-executable"
	}

	argsStr := ""
	if customArgs != "" {
		argsStr = " " + customArgs
	}

	categoriesStr := resolveCategories(customCategories)
	wmClass := sanitizeWMClass(appName)

	content := fmt.Sprintf(desktopTemplate,
		appName, description, absAppImage, argsStr, iconPath, categoriesStr, wmClass)

	menuEntry := filepath.Join(m.ApplicationsDir, sanitizeFilename(appName)+".desktop")

	if m.DryRun {
		return content, nil
	}

	if err := os.MkdirAll(m.ApplicationsDir, 0755); err != nil {
		return "", fmt.Errorf("Konnte Verzeichnis nicht erstellen: %v", err)
	}

	if err := os.WriteFile(menuEntry, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("Konnte Menue-Eintrag nicht schreiben: %v", err)
	}

	if m.DesktopIcon {
		if err := os.MkdirAll(m.DesktopPath, 0755); err != nil {
			return "", fmt.Errorf("Konnte Desktop-Verzeichnis nicht erstellen: %v", err)
		}
		desktopEntry := filepath.Join(m.DesktopPath, sanitizeFilename(appName)+".desktop")
		if err := os.WriteFile(desktopEntry, []byte(content), 0755); err != nil {
			return "", fmt.Errorf("Konnte Desktop-Eintrag nicht schreiben: %v", err)
		}
		_ = exec.Command("gio", "set", desktopEntry, "metadata::trusted", "true").Run()
	}

	_ = exec.Command("update-desktop-database", m.ApplicationsDir).Run()
	return appName, nil
}

func runInstall(appImagePath, customName, customDesc, customIcon, customArgs,
	customCategories *string, desktop, noExtractIcon, dryRun *bool) {

	m := NewManager()
	m.DryRun = *dryRun
	m.NoExtractIcon = *noExtractIcon
	m.DesktopIcon = *desktop

	if *appImagePath == "" {
		fmt.Fprintln(os.Stderr, "Fehler: --appimage ist erforderlich.")
		installUsage()
		os.Exit(1)
	}

	fmt.Printf("\n\U0001f4e6 AppImage-Installer\n")
	fmt.Printf("======================================\n")

	res, err := m.Install(*appImagePath, *customName, *customDesc, *customIcon, *customArgs, *customCategories)
	if err != nil {
		fatalf("%v", err)
	}

	if *dryRun {
		fmt.Println("-- Dry-Run ----------------------------------")
		fmt.Println("Wuerde folgende .desktop-Datei schreiben:")
		fmt.Println()
		fmt.Print(res)
		fmt.Println("\nKeine Aenderungen vorgenommen (--dry-run).")
		return
	}

	fmt.Printf("\n\"%s\" wurde erfolgreich installiert!\n\n", res)
}

func launchGUI() {
	app := NewApp()
	
	// Start a local server to serve embedded files
	subFS, err := fs.Sub(frontendAssets, "frontend")
	if err != nil {
		fatalf("Konnte Frontend-Dateien nicht finden: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fatalf("Konnte keinen lokalen Port oeffnen: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d/index.html", port)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(subFS)))
	
	// Server local icons to bypass "Not allowed to load local resource"
	m := NewManager()
	mux.Handle("/icons/", http.StripPrefix("/icons/", http.FileServer(http.Dir(m.IconsDir))))

	server := &http.Server{
		Handler: mux,
	}
	go server.Serve(listener)

	w := webview.New(true)
	defer w.Destroy()
	app.w = w

	w.SetTitle("AppInstaller")
	w.SetSize(900, 650, webview.HintNone)

	// Bind Go methods to JS
	w.Bind("GetInstalledApps", app.GetInstalledApps)
	w.Bind("SelectAppImage", app.SelectAppImage)
	w.Bind("InstallApp", app.Install)
	w.Bind("UninstallApp", app.Uninstall)
	w.Bind("GetConfig", app.GetConfig)
	w.Bind("SaveConfig", app.SaveConfig)
	w.Bind("SelectFolder", app.SelectFolder)
	w.Bind("CheckForUpdates", app.CheckForUpdates)

	w.Navigate(url)
	w.Run()
}
// ---------------------------------------------------------------------------
// uninstall
// ---------------------------------------------------------------------------

type installedEntry struct {
	Name        string
	DesktopFile string
	DesktopPath string
	IconPath    string
}

func (m *Manager) FindInstalledEntries() []installedEntry {
	matches, _ := filepath.Glob(filepath.Join(m.ApplicationsDir, "*.desktop"))

	var entries []installedEntry
	for _, f := range matches {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		content := string(data)
		isOurs := strings.Contains(content, "appimage-install") ||
			strings.Contains(strings.ToLower(content), ".appimage")
		if !isOurs {
			continue
		}

		name := ""
		iconPath := ""
		for _, line := range strings.Split(content, "\n") {
			if strings.HasPrefix(line, "Name=") {
				name = strings.TrimPrefix(line, "Name=")
			}
			if strings.HasPrefix(line, "Icon=") {
				iconPath = strings.TrimPrefix(line, "Icon=")
			}
		}
		if name == "" {
			name = strings.TrimSuffix(filepath.Base(f), ".desktop")
		}

		desktopPath := filepath.Join(m.DesktopPath, filepath.Base(f))
		if !fileExists(desktopPath) {
			desktopPath = ""
		}

		if !strings.HasPrefix(iconPath, m.IconsDir) {
			iconPath = ""
		}

		entries = append(entries, installedEntry{
			Name:        name,
			DesktopFile: f,
			DesktopPath: desktopPath,
			IconPath:    iconPath,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return entries
}

func runUninstall(nameFlag *string, listFlag, dryRun *bool) {
	m := NewManager()
	m.DryRun = *dryRun
	entries := m.FindInstalledEntries()

	if *listFlag {
		if len(entries) == 0 {
			fmt.Println("\nKeine per appimage-install installierten Apps gefunden.")
			return
		}
		fmt.Printf("\nInstallierte AppImages (%d)\n", len(entries))
		fmt.Println("======================================")
		for i, e := range entries {
			extra := ""
			if e.DesktopPath != "" {
				extra = " [+Desktop]"
			}
			fmt.Printf("  %2d. %s%s\n", i+1, e.Name, extra)
			fmt.Printf("      %s\n", e.DesktopFile)
		}
		fmt.Println()
		return
	}

	var target *installedEntry

	if *nameFlag != "" {
		for i := range entries {
			if strings.EqualFold(entries[i].Name, *nameFlag) {
				target = &entries[i]
				break
			}
		}
		if target == nil {
			var idxMatches []int
			for i := range entries {
				if strings.Contains(strings.ToLower(entries[i].Name), strings.ToLower(*nameFlag)) {
					idxMatches = append(idxMatches, i)
				}
			}
			if len(idxMatches) == 1 {
				target = &entries[idxMatches[0]]
			} else if len(idxMatches) > 1 {
				fmt.Printf("Mehrere Treffer fuer \"%s\":\n\n", *nameFlag)
				for _, idx := range idxMatches {
					fmt.Printf("   - %s\n", entries[idx].Name)
				}
				fmt.Println("\nBitte den Namen genauer angeben.")
				os.Exit(1)
			}
		}
		if target == nil {
			fmt.Printf("Keine installierte App gefunden: \"%s\"\n", *nameFlag)
			fmt.Println("Tipp: appimage-install uninstall --list")
			os.Exit(1)
		}
	} else {
		if len(entries) == 0 {
			fmt.Println("\nKeine per appimage-install installierten Apps gefunden.")
			return
		}
		fmt.Println("\nInstallierte AppImages - welche soll deinstalliert werden?")
		fmt.Println("======================================")
		for i, e := range entries {
			extra := ""
			if e.DesktopPath != "" {
				extra = " [+Desktop]"
			}
			fmt.Printf("  %2d. %s%s\n", i+1, e.Name, extra)
		}
		fmt.Println("   0. Abbrechen")
		fmt.Println()
		fmt.Print("Auswahl: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
		if input == "0" || input == "" {
			fmt.Println("Abgebrochen.")
			return
		}
		idx, err := strconv.Atoi(input)
		if err != nil || idx < 1 || idx > len(entries) {
			fmt.Println("Ungueltige Auswahl.")
			os.Exit(1)
		}
		target = &entries[idx-1]
	}

	fmt.Printf("\nDeinstallation: %s\n", target.Name)
	fmt.Println("======================================")
	fmt.Printf("  Menue-Eintrag : %s\n", target.DesktopFile)
	if target.DesktopPath != "" {
		fmt.Printf("  Desktop       : %s\n", target.DesktopPath)
	}
	if target.IconPath != "" {
		fmt.Printf("  Icon          : %s\n", target.IconPath)
	}
	fmt.Println()

	if *dryRun {
		fmt.Println("-- Dry-Run: Keine Dateien geloescht. --")
		return
	}

	fmt.Print("Wirklich deinstallieren? [j/N] ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if answer != "j" && answer != "ja" && answer != "y" && answer != "yes" {
		fmt.Println("Abgebrochen.")
		return
	}

	if err := m.Uninstall(*target); err != nil {
		fmt.Printf("WARN: %v\n", err)
	} else {
		fmt.Printf("\n\"%s\" wurde erfolgreich deinstalliert!\n\n", target.Name)
	}
}

// ---------------------------------------------------------------------------
// Icon-Extraktion
// ---------------------------------------------------------------------------

func (m *Manager) extractIcon(appImagePath, appName string) string {
	if m.DryRun {
		return ""
	}

	if err := os.MkdirAll(m.IconsDir, 0755); err != nil {
		return ""
	}

	tmpDir, err := os.MkdirTemp("", "appimage-extract-*")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command(appImagePath, "--appimage-extract")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		return ""
	}

	squashfsRoot := filepath.Join(tmpDir, "squashfs-root")

	type iconCandidate struct {
		path  string
		score int
	}
	var candidates []iconCandidate

	iconSearchPaths := []struct {
		pattern string
		base    int
	}{
		{filepath.Join(squashfsRoot, "usr", "share", "icons", "*", "*x*", "*", "*.png"), 100},
		{filepath.Join(squashfsRoot, "usr", "share", "icons", "*", "scalable", "*", "*.svg"), 60},
		{filepath.Join(squashfsRoot, "usr", "share", "pixmaps", "*.png"), 50},
		{filepath.Join(squashfsRoot, "usr", "share", "pixmaps", "*.svg"), 40},
		{filepath.Join(squashfsRoot, "*.png"), 30},
		{filepath.Join(squashfsRoot, "*.svg"), 20},
	}

	for _, sp := range iconSearchPaths {
		matches, _ := filepath.Glob(sp.pattern)
		for _, match := range matches {
			score := sp.base
			for _, sizeBonus := range []struct {
				s     string
				bonus int
			}{
				{"512", 50}, {"256", 40}, {"128", 30}, {"96", 20}, {"64", 10}, {"48", 5},
			} {
				if strings.Contains(match, sizeBonus.s) {
					score += sizeBonus.bonus
					break
				}
			}
			candidates = append(candidates, iconCandidate{path: match, score: score})
		}
	}

	dirIcon := filepath.Join(squashfsRoot, ".DirIcon")
	if fileExists(dirIcon) {
		candidates = append(candidates, iconCandidate{path: dirIcon, score: 5})
	}

	if len(candidates) == 0 {
		return ""
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	best := candidates[0].path

	ext := filepath.Ext(best)
	if ext == "" {
		ext = ".png"
	}
	destIcon := filepath.Join(m.IconsDir, sanitizeFilename(appName)+ext)

	data, err := os.ReadFile(best)
	if err != nil {
		return ""
	}
	if err := os.WriteFile(destIcon, data, 0644); err != nil {
		return ""
	}

	return destIcon
}

func (m *Manager) Uninstall(entry installedEntry) error {
	var errs []string

	if err := os.Remove(entry.DesktopFile); err != nil {
		errs = append(errs, fmt.Sprintf("Menue-Eintrag: %v", err))
	}

	if entry.DesktopPath != "" {
		if err := os.Remove(entry.DesktopPath); err != nil {
			errs = append(errs, fmt.Sprintf("Desktop-Eintrag: %v", err))
		}
	}

	if entry.IconPath != "" {
		if err := os.Remove(entry.IconPath); err != nil {
			errs = append(errs, fmt.Sprintf("Icon: %v", err))
		}
	}

	_ = exec.Command("update-desktop-database", m.ApplicationsDir).Run()

	if len(errs) > 0 {
		return fmt.Errorf("Fehler beim Loeschen einiger Dateien: %s", strings.Join(errs, ", "))
	}
	return nil
}

// ---------------------------------------------------------------------------
// Kategorien
// ---------------------------------------------------------------------------

func resolveCategories(input string) string {
	if input == "" {
		return "Utility;"
	}
	var resolved []string
	warned := false
	for _, p := range strings.Split(input, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if canonical, ok := validCategories[strings.ToLower(p)]; ok {
			resolved = append(resolved, canonical)
			continue
		}
		found := false
		for _, v := range validCategories {
			if strings.EqualFold(v, p) {
				resolved = append(resolved, v)
				found = true
				break
			}
		}
		if !found {
			if !warned {
				fmt.Printf("WARN: Unbekannte Kategorie(n) werden direkt uebernommen.\n")
				fmt.Printf("      Tipp: appimage-install install --list-categories\n")
				warned = true
			}
			resolved = append(resolved, p)
		}
	}
	if len(resolved) == 0 {
		return "Utility;"
	}
	return strings.Join(resolved, ";") + ";"
}

func printCategories() {
	mainCats := []string{
		"AudioVideo", "Audio", "Video", "Development", "Education",
		"Game", "Graphics", "Network", "Office", "Science",
		"Settings", "System", "Utility",
	}
	mainSet := map[string]bool{}
	for _, c := range mainCats {
		mainSet[c] = true
	}
	var additional []string
	for _, v := range validCategories {
		if !mainSet[v] {
			additional = append(additional, v)
		}
	}
	sort.Strings(additional)

	fmt.Println("\nVerfuegbare Kategorien")
	fmt.Println("======================================")
	fmt.Println("\n-- Hauptkategorien --")
	for _, c := range mainCats {
		fmt.Printf("   %-20s\n", c)
	}
	fmt.Println("\n-- Zusatzkategorien --")
	cols := 3
	for i, c := range additional {
		fmt.Printf("   %-22s", c)
		if (i+1)%cols == 0 {
			fmt.Println()
		}
	}
	if len(additional)%cols != 0 {
		fmt.Println()
	}
	fmt.Println(`
Verwendung: --categories "Network,Chat"
           Mehrere Kategorien kommagetrennt angeben.
`)
}

// ---------------------------------------------------------------------------
// Hilfsfunktionen
// ---------------------------------------------------------------------------

func sanitizeName(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_'
	})
	var clean []string
	for _, p := range parts {
		if len(p) > 0 && (p[0] == 'v' || p[0] == 'V') {
			if len(p) > 1 && p[1] >= '0' && p[1] <= '9' {
				break
			}
		}
		allDigitsOrDots := true
		for _, c := range p {
			if !((c >= '0' && c <= '9') || c == '.') {
				allDigitsOrDots = false
				break
			}
		}
		if allDigitsOrDots && len(p) > 0 {
			break
		}
		clean = append(clean, p)
	}
	if len(clean) == 0 {
		return name
	}
	return strings.Join(clean, " ")
}

func sanitizeFilename(name string) string {
	r := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-")
	return strings.ToLower(r.Replace(name))
}

// sanitizeWMClass erzeugt einen StartupWMClass-Wert.
// Viele Electron/Qt-Apps setzen WMClass auf den ersten Token des App-Namens.
func sanitizeWMClass(name string) string {
	parts := strings.Fields(name)
	if len(parts) > 0 {
		name = parts[0]
	}
	r := strings.NewReplacer(" ", "", "/", "", "\\", "", ":", "")
	return r.Replace(name)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "FEHLER: "+format+"\n", args...)
	os.Exit(1)
}

// ---------------------------------------------------------------------------
// Hilfe-Texte
// ---------------------------------------------------------------------------

func globalUsage() {
	fmt.Fprint(os.Stderr, `
AppImage-Installer - AppImages ins Startmenue und/oder Desktop installieren

VERWENDUNG:
  appinstall <befehl> [Optionen]

BEFEHLE:
  install      AppImage installieren
  uninstall    AppImage deinstallieren (auch: remove)
  help         Diese Hilfe anzeigen

BEISPIELE:
  appinstall install --appimage ~/Downloads/MyApp.AppImage
  appinstall uninstall --list
  appinstall uninstall --name "MyApp"

Weitere Hilfe:
  appinstall install --help
  appinstall uninstall --help

`)
}

func installUsage() {
	fmt.Fprint(os.Stderr, `
VERWENDUNG:
  appimage-install install --appimage <datei.AppImage> [Optionen]

OPTIONEN:
  --appimage <pfad>          Pfad zur AppImage-Datei (erforderlich)
  --name <n>              Benutzerdefinierter Anwendungsname
  --description <text>       Kurzbeschreibung der Anwendung
  --icon <pfad>              Benutzerdefiniertes Icon (.png, .svg, .xpm)
  --args <argumente>         Zusaetzliche Start-Argumente, z.B. "--no-sandbox"
  --categories <kat,...>     Kommagetrennte Kategorien, z.B. "Network,Chat"
  --list-categories          Alle verfuegbaren Kategorien anzeigen
  --desktop                  Auch auf dem Desktop installieren
  --no-extract-icon          Icon-Extraktion aus dem AppImage ueberspringen
  --dry-run                  Vorschau anzeigen, ohne Dateien zu schreiben

BEISPIELE:
  appimage-install install --appimage ~/Downloads/MyApp-1.2.3.AppImage

  appimage-install install --appimage ~/Downloads/MyApp.AppImage \
    --name "Meine App" --description "Meine tolle Anwendung" \
    --categories "Development,TextEditor" --desktop

  appimage-install install --appimage ~/Downloads/MyApp.AppImage \
    --icon ~/Bilder/myapp.png --args "--no-sandbox"

  appimage-install install --list-categories
  appimage-install install --appimage ~/Downloads/MyApp.AppImage --dry-run

`)
}

func uninstallUsage() {
	fmt.Fprint(os.Stderr, `
VERWENDUNG:
  appimage-install uninstall [Optionen]

OPTIONEN:
  --name <n>    Name der zu deinstallierenden App (auch Teilstring)
  --list           Alle installierten Apps auflisten
  --dry-run        Vorschau anzeigen, ohne Dateien zu loeschen

BEISPIELE:
  # Interaktiv auswaehlen:
  appimage-install uninstall

  # Direkt per Name:
  appimage-install uninstall --name "MyApp"

  # Alle installierten Apps anzeigen:
  appimage-install uninstall --list

  # Vorschau:
  appimage-install uninstall --name "MyApp" --dry-run

`)
}
